import os
from concurrent import futures
from datetime import datetime

import grpc

import analytics_pb2 as pb
import analytics_pb2_grpc as pb_grpc
from mappers import ReportDataMapper
from report_generator import render_any_report


class AnalyticsService(pb_grpc.AnalyticsServiceServicer):
    def GenerateReport(self, request, context):
        started_at = datetime.now()
        session_id = request.session_id
        report_format = request.format

        print(f"\n[{started_at.strftime('%H:%M:%S')}] >>> [GENERATE REPORT]", flush=True)
        print(f"    SessionID: {session_id} | Format: {report_format}", flush=True)

        try:
            report_data = ReportDataMapper.map_request_to_data(request)
            template_bytes, file_name, content_type = self._resolve_template(report_format)
            file_bytes = render_any_report(template_bytes, report_data, file_name)

            duration = (datetime.now() - started_at).total_seconds()
            print(
                f"<<< [SUCCESS] Report generated in {duration:.2f}s. Size: {len(file_bytes)} bytes",
                flush=True,
            )

            return pb.GenerateReportResponse(
                file_content=file_bytes,
                content_type=content_type,
                suggested_filename=file_name,
            )
        except Exception as exc:
            print(f"[!!!] SERVER ERROR: {exc}", flush=True)
            import traceback

            traceback.print_exc()
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(str(exc))
            return pb.GenerateReportResponse()

    def _resolve_template(self, report_format):
        mapping = {
            pb.REPORT_FORMAT_HTML: (
                "report_template.html",
                "report.html",
                "text/html; charset=utf-8",
            ),
            pb.REPORT_FORMAT_PSYCHO_HTML: (
                "report_for_psychologist.html",
                "psychologist-report.html",
                "text/html; charset=utf-8",
            ),
            pb.REPORT_FORMAT_PSYCHO_DOCX: (
                "template_psychologist.docx",
                "psychologist-report.docx",
                "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            ),
        }

        path, name, mime = mapping.get(
            report_format,
            (
                "template_client01.docx",
                "client-report.docx",
                "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
            ),
        )

        with open(path, "rb") as file:
            return file.read(), name, mime


def serve():
    port = os.getenv("PORT", "50051")
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    pb_grpc.add_AnalyticsServiceServicer_to_server(AnalyticsService(), server)
    server.add_insecure_port(f"[::]:{port}")

    print("=" * 50, flush=True)
    print(f"PYTHON ANALYTICS SERVICE RUNNING ON PORT {port}", flush=True)
    print("=" * 50, flush=True)

    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()
