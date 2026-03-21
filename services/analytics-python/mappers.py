import json
from datetime import datetime, timedelta

class ReportDataMapper:
    @staticmethod
    def weight_to_text(weight):
        mapping = {
            5: "Экспертные знания / Знаю отлично",
            4: "Хорошие знания / Знаю теорию и практику",
            3: "Базовые знания / Знаю теорию",
            2: "Поверхностные знания",
            1: "Не владею / Не знаю"
        }
        return mapping.get(int(weight), "Нет данных")

    @classmethod
    def map_request_to_data(cls, request):
        # 0. Парсинг метаданных
        meta = json.loads(request.client_metadata_json) if request.client_metadata_json else {}
        categories = ['Аналитика', 'Код', 'Дизайн', 'Тесты', 'Менеджмент']
        cat_map = {c: [] for c in categories}

        q_answers = []
        critical_answers = []

        # 1. Сбор сырых данных
        for i, r in enumerate(request.responses, 1):
            cat_map[r.category_tag].append(r.selected_weight)
            answer_text = r.raw_text if r.raw_text else cls.weight_to_text(r.selected_weight)

            ans_obj = {
                "id": i, "question": r.question_text, "answer": answer_text,
                "item_score": r.selected_weight, "text": r.question_text,
                "value": answer_text, "time": 15
            }
            q_answers.append(ans_obj)
            if r.selected_weight in [1, 5]:
                critical_answers.append(ans_obj)

        # 2. Расчет факторов
        factors = []
        for i, cat in enumerate(categories, 1):
            avg = sum(cat_map[cat]) / len(cat_map[cat]) if cat_map[cat] else 0.0
            factors.append({
                "no": i, "name": cat, "score": round(avg, 1),
                "level": "Высокий" if avg >= 4 else "Средний" if avg >= 2.5 else "Низкий",
                "profile_type": "Лидирующий" if avg >= 4.2 else "Стабильный",
                "psych_interpretation": f"Демонстрирует устойчивый навык в категории {cat}."
            })

        top = sorted(factors, key=lambda x: x['score'], reverse=True)

        # 3. Формирование финального словаря
        return {
            "user_name": f"{meta.get('last_name', '')} {meta.get('first_name', '')}".strip() or "Кандидат",
            "session_id": request.session_id,
            "profile_id": f"PR-{request.session_id[:5].upper()}",
            "date": datetime.now().strftime("%d.%m.%Y"),
            "duration": meta.get("duration", "15:00"),
            "main_sphere": meta.get("main_sphere", "Разработка"),
            "notes": meta.get("notes", "Профиль сбалансирован."),
            "factors": factors,
            "q_answers": q_answers,
            "critical_answers": critical_answers,
            "labels": categories,
            "scores": [f['score'] for f in factors],
            "job_1": "System Architect",
            "job_2": "Senior Backend Developer",
            "job_3": "Data Engineer",
            "strong_skill_1_name": top[0]['name'],
            "strong_skill_1_score": int(top[0]['score'] * 20),
            "strong_skill_1_desc": f"Категория {top[0]['name']} — ведущая компетенция.",
            "strong_skill_1_dev": "Рекомендуется участие в архитектурных комитетах.",
            "strong_skill_2_name": top[1]['name'],
            "strong_skill_2_score": int(top[1]['score'] * 20),
            "strong_skill_2_desc": f"Навыки {top[1]['name']} на хорошем уровне.",
            "strong_skill_2_dev": "Фокусируйтесь на наставничестве.",
            "project_focus_1": "Масштабируемые системы",
            "project_example_1": "Разработка высоконагруженных API",
            "next_test_date": (datetime.now() + timedelta(days=180)).strftime("%d.%m.%Y"),
            "support_email": "support@profdnk.ru"
        }