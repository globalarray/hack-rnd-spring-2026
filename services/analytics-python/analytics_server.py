import grpc
import json
from concurrent import futures
from datetime import datetime, timedelta
import test_engine_pb2 as pb
import test_engine_pb2_grpc as pb_grpc
from report_generator import render_any_report


class AnalyticsService(pb_grpc.AnalyticsServiceServicer):
    def GenerateReport(self, request, context):
        try:
            meta = json.loads(request.client_metadata_json) if request.client_metadata_json else {}
            categories = ['Аналитика', 'Код', 'Дизайн', 'Тесты', 'Менеджмент']
            cat_map = {c: [] for c in categories}

            q_answers = []
            critical_answers = []

            # 1. Сбор сырых данных
            for i, r in enumerate(request.responses, 1):
                cat_map[r.category_tag].append(r.selected_weight)

                # Текст ответа на основе веса (для 2-й ячейки в таблице психолога)
                answer_text = r.raw_text if r.raw_text else self._weight_to_text(r.selected_weight)

                ans_obj = {
                    "id": i,
                    "question": r.question_text,  # {{ item.question }}
                    "answer": answer_text,  # {{ item.answer }} -> "Знаю теорию" и т.д.
                    "item_score": r.selected_weight,  # {{ item.item_score }} -> 5.0
                    "text": r.question_text,  # Для блока критических точек
                    "value": answer_text,
                    "time": 15  # Условно
                }
                q_answers.append(ans_obj)

                if r.selected_weight in [1, 5]:
                    critical_answers.append(ans_obj)

            # 2. Расчет факторов
            factors = []
            for i, cat in enumerate(categories, 1):
                avg = sum(cat_map[cat]) / len(cat_map[cat]) if cat_map[cat] else 0.0
                factors.append({
                    "no": i,
                    "name": cat,
                    "score": round(avg, 1),
                    "level": "Высокий" if avg >= 4 else "Средний" if avg >= 2.5 else "Низкий",
                    "profile_type": "Лидирующий" if avg >= 4.2 else "Стабильный",
                    "psych_interpretation": f"Демонстрирует устойчивый навык в категории {cat}."
                })

            # 3. Топ-скиллы и профессии для пользователя
            top = sorted(factors, key=lambda x: x['score'], reverse=True)

            full_data = {
                "user_name": f"{meta.get('last_name', '')} {meta.get('first_name', '')}".strip() or "Кандидат",
                "session_id": request.session_id,
                "profile_id": f"PR-{request.session_id[:5].upper()}",
                "date": datetime.now().strftime("%d.%m.%Y"),
                "duration": meta.get("duration", "15:00"),
                "main_sphere": meta.get("main_sphere", "Разработка"),
                "notes": meta.get("notes", "Профиль сбалансирован, выражены аналитические способности."),

                "factors": factors,
                "q_answers": q_answers,
                "critical_answers": critical_answers,
                "labels": categories,
                "scores": [f['score'] for f in factors],

                # Топ-3 профессии (теги {{ job_1 }} и т.д.)
                "job_1": "System Architect",
                "job_2": "Senior Backend Developer",
                "job_3": "Data Engineer",

                # Скиллы (теги {{ strong_skill_1_name }} и т.д.)
                "strong_skill_1_name": top[0]['name'],
                "strong_skill_1_score": int(top[0]['score'] * 20),
                "strong_skill_1_desc": f"Категория {top[0]['name']} является вашей ведущей компетенцией. Вы способны решать задачи высокой сложности в этом векторе.",
                "strong_skill_1_dev": "Рекомендуется расширение стека смежных технологий и участие в архитектурных комитетах.",

                "strong_skill_2_name": top[1]['name'],
                "strong_skill_2_score": int(top[1]['score'] * 20),
                "strong_skill_2_desc": f"Навыки в области {top[1]['name']} позволяют эффективно поддерживать рабочие процессы.",
                "strong_skill_2_dev": "Фокусируйтесь на оптимизации текущих решений и наставничестве младших коллег.",

                # Проекты (теги {{ project_focus_1 }} и т.д.)
                "project_focus_1": "Масштабируемые системы",
                "project_example_1": "Разработка высоконагруженных API",
                "project_focus_2": "Безопасность данных",
                "project_example_2": "Внедрение протоколов шифрования",
                "project_focus_3": "Оптимизация запросов",
                "project_example_3": "Тюнинг производительности БД",

                "next_test_date": (datetime.now() + timedelta(days=180)).strftime("%d.%m.%Y"),
                "support_email": "support@profdnk.ru"
            }

            is_html = b'<html' in request.template_content.lower()
            file_bytes = render_any_report(request.template_content, full_data, "res.html" if is_html else "res.docx")
            return pb.GenerateReportResponse(file_content=file_bytes)

        except Exception as e:
            print(f"Server error: {e}")
            context.set_code(grpc.StatusCode.INTERNAL)
            return pb.GenerateReportResponse()

    def _weight_to_text(self, weight):
        mapping = {
            5: "Экспертные знания / Знаю отлично",
            4: "Хорошие знания / Знаю теорию и практику",
            3: "Базовые знания / Знаю теорию",
            2: "Поверхностные знания",
            1: "Не владею / Не знаю"
        }
        return mapping.get(int(weight), "Нет данных")


def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    pb_grpc.add_AnalyticsServiceServicer_to_server(AnalyticsService(), server)
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()


if __name__ == '__main__':
    serve()
