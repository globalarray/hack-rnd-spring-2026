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
        meta = cls._extract_metadata(request)
        category_scores = {}

        q_answers = []
        critical_answers = []

        for i, r in enumerate(request.responses, 1):
            category_key = cls._normalize_category(r.category_tag)
            category_scores.setdefault(category_key, []).append(r.selected_weight)
            answer_text = r.raw_text if r.raw_text else cls.weight_to_text(r.selected_weight)

            ans_obj = {
                "id": i,
                "question": r.question_text,
                "answer": answer_text,
                "item_score": r.selected_weight,
                "text": r.question_text,
                "value": answer_text,
                "time": 15,
            }
            q_answers.append(ans_obj)
            if r.selected_weight in [1, 5]:
                critical_answers.append(ans_obj)

        if not category_scores:
            category_scores["General"] = [0.0]

        factors = []
        for i, (category_name, weights) in enumerate(category_scores.items(), 1):
            avg = sum(weights) / len(weights) if weights else 0.0
            factors.append({
                "no": i,
                "name": category_name,
                "score": round(avg, 1),
                "raw_score": round(avg, 1),
                "level": "Высокий" if avg >= 4 else "Средний" if avg >= 2.5 else "Низкий",
                "profile_type": "Лидирующий" if avg >= 4.2 else "Стабильный",
                "accentuation": cls._accentuation(avg),
                "psych_interpretation": f"Демонстрирует устойчивый навык в категории {category_name}.",
                "psych_comment": cls._psych_comment(category_name, avg),
            })

        top = sorted(factors, key=lambda x: x['score'], reverse=True)
        top_1 = top[0]
        top_2 = top[1] if len(top) > 1 else top[0]
        labels = [factor["name"] for factor in factors]
        scores = [factor["score"] for factor in factors]

        return {
            "user_name": cls._resolve_user_name(meta),
            "session_id": request.session_id,
            "profile_id": f"PR-{request.session_id[:5].upper()}",
            "date": datetime.now().strftime("%d.%m.%Y"),
            "duration": meta.get("duration", "15:00"),
            "main_sphere": meta.get("main_sphere", "Разработка"),
            "notes": meta.get("notes", "Профиль сбалансирован."),
            "sincerity_score": meta.get("sincerity_score", 92),
            "factors": factors,
            "q_answers": q_answers,
            "critical_answers": critical_answers,
            "labels": labels,
            "scores": scores,
            "job_1": "System Architect",
            "job_2": "Senior Backend Developer",
            "job_3": "Data Engineer",
            "strong_skill_1_name": top_1['name'],
            "strong_skill_1_score": int(top_1['score'] * 20),
            "strong_skill_1_desc": f"Категория {top_1['name']} — ведущая компетенция.",
            "strong_skill_1_dev": "Рекомендуется участие в архитектурных комитетах.",
            "strong_skill_2_name": top_2['name'],
            "strong_skill_2_score": int(top_2['score'] * 20),
            "strong_skill_2_desc": f"Навыки {top_2['name']} на хорошем уровне.",
            "strong_skill_2_dev": "Фокусируйтесь на наставничестве.",
            "project_focus_1": "Масштабируемые системы",
            "project_example_1": "Разработка высоконагруженных API",
            "next_test_date": (datetime.now() + timedelta(days=180)).strftime("%d.%m.%Y"),
            "support_email": "support@profdnk.ru"
        }

    @classmethod
    def map_go_data_to_report(cls, request, _session_id=None):
        return cls.map_request_to_data(request)

    @staticmethod
    def _extract_metadata(request):
        raw_metadata = getattr(request, "client_metadata_json", "")
        if not raw_metadata:
            return {}

        try:
            import json
            return json.loads(raw_metadata)
        except Exception:
            return {}

    @staticmethod
    def _normalize_category(category_tag):
        raw = (category_tag or "").strip()
        if not raw:
            return "General"

        raw = raw.replace("-", " ").replace("_", " ")
        return " ".join(part.capitalize() for part in raw.split())

    @staticmethod
    def _resolve_user_name(meta):
        full_name = str(
            meta.get("full_name")
            or meta.get("fullName")
            or meta.get("fio")
            or meta.get("name")
            or ""
        ).strip()
        if full_name:
            return full_name

        first_name = str(meta.get("first_name") or meta.get("firstName") or "").strip()
        last_name = str(meta.get("last_name") or meta.get("lastName") or "").strip()
        combined = f"{last_name} {first_name}".strip()
        return combined or "Кандидат"

    @staticmethod
    def _accentuation(score):
        if score >= 4.2:
            return "Выражено"
        if score >= 2.5:
            return "Норма"
        return "Зона развития"

    @staticmethod
    def _psych_comment(category_name, score):
        if score >= 4.2:
            prefix = "Сильная выраженность"
        elif score >= 2.5:
            prefix = "Умеренная выраженность"
        else:
            prefix = "Низкая выраженность"
        return f"{prefix} по шкале «{category_name}»."
