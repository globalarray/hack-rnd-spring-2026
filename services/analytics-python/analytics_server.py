import os
import grpc
from concurrent import futures
from datetime import datetime

import test_engine_pb2 as pb
import test_engine_pb2_grpc as pb_grpc
import target_service_pb2 as target_pb
import target_service_pb2_grpc as target_pb_grpc

from mappers import ReportDataMapper
from report_generator import render_any_report


class AnalyticsService(pb_grpc.AnalyticsServiceServicer):
    def __init__(self, go_stub):
        self.go_stub = go_stub

    def GenerateReport(self, request, context):
        try:
            # 1. Стучимся в Go-сервис (например, за логами или статусом)
            try:
                self.go_stub.SomeMethod(target_pb.SomeRequest(payload=request.session_id), timeout=2)
            except grpc.RpcError as e:
                print(f"Go service error (skipped): {e.details()}")

            # 2. Маппинг данных (вынесено в отдельный модуль)
            full_data = ReportDataMapper.map_request_to_data(request)

            # 3. Рендеринг отчета
            is_html = b'<html' in request.template_content.lower()
            file_bytes = render_any_report(
                request.template_content,
                full_data,
                "res.html" if is_html else "res.docx"
            )

            return pb.GenerateReportResponse(file_content=file_bytes)

        except Exception as e:
            print(f"Critical server error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(e))
            return pb.GenerateReportResponse()


def serve():
    cert_path = os.getenv('CERTS_PATH', '/etc/certs')
    go_addr = os.getenv('GO_SERVICE_ADDR', 'go-service:50036')

    # Загрузка сертификатов
    try:
        with open(os.path.join(cert_path, 'server.key'), 'rb') as f:
            pk = f.read()
        with open(os.path.join(cert_path, 'server.crt'), 'rb') as f:
            cert = f.read()
        with open(os.path.join(cert_path, 'ca.crt'), 'rb') as f:
            ca = f.read()
    except FileNotFoundError as e:
        print(f"Certs not found: {e}")
        return

    # Настройка клиента для Go
    cl_creds = grpc.ssl_channel_credentials(ca, pk, cert)
    go_chan = grpc.secure_channel(go_addr, cl_creds)
    go_stub = target_pb_grpc.TargetServiceStub(go_chan)

    # Настройка сервера
    srv_creds = grpc.ssl_server_credentials([(pk, cert)], ca, True)
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    pb_grpc.add_AnalyticsServiceServicer_to_server(AnalyticsService(go_stub), server)

    server.add_secure_port('[::]:50051', srv_creds)
    print(f"--- Analytics Service Started (mTLS) ---\nListening on :50051\nTarget Go: {go_addr}")

    server.start()
    server.wait_for_termination()


if __name__ == '__main__':
    serve()