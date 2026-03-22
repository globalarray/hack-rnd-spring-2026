import type { ReportPreview, SessionAnalytics, SurveyRecord } from "./types";

function topCategories(session: SessionAnalytics) {
  const counters = new Map<string, number>();

  session.responses.forEach((response) => {
    if (!response.categoryTag) {
      return;
    }

    response.categoryTag
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean)
      .forEach((tag) => {
        counters.set(tag, (counters.get(tag) ?? 0) + (response.selectedWeight ?? 1));
      });
  });

  return [...counters.entries()].sort((left, right) => right[1] - left[1]);
}

function humanizeCategory(tag: string) {
  switch (tag) {
    case "people":
      return "работа с людьми";
    case "analysis":
      return "аналитика и исследование";
    case "systems":
      return "системное мышление";
    default:
      return tag;
  }
}

export function buildReportPreview(survey: SurveyRecord, session: SessionAnalytics): ReportPreview {
  const categories = topCategories(session);
  const strongest = categories[0]?.[0];
  const second = categories[1]?.[0];
  const rawResponses = session.responses.filter((item) => item.rawText).map((item) => item.rawText as string);

  return {
    title: `Отчет по методике «${survey.title}»`,
    subtitle: `Сессия ${session.sessionId.slice(0, 8)} для ${String(session.clientMetadata.fullName ?? "участника")}`,
    summary: [
      strongest
        ? `Наиболее ярко проявилась ориентация на ${humanizeCategory(strongest)}.`
        : "Категориальные ответы пока не доминируют, поэтому лучше опираться на качественный разбор.",
      second ? `Второй заметный вектор: ${humanizeCategory(second)}.` : "Второй устойчивый вектор пока не выделился.",
      `Всего проанализировано ${session.responses.length} ответов.`
    ],
    strengths: [
      strongest ? `Участнику подходит среда, где есть ${humanizeCategory(strongest)}.` : "Участник достаточно гибко распределяет интересы.",
      rawResponses[0]
        ? `В открытом ответе проявился запрос: «${rawResponses[0]}».`
        : "Открытые вопросы не дали развернутых формулировок, стоит уточнить мотивацию на интервью."
    ],
    developmentPoints: [
      "Рекомендуется обсудить условия, в которых участник быстрее принимает решения.",
      "Полезно проверить устойчивость интересов на реальных кейсах или мини-практике."
    ],
    responseHighlights: session.responses.map((response) => {
      if (response.rawText) {
        return `${response.questionText}: ${response.rawText}`;
      }

      return `${response.questionText}: вес ${response.selectedWeight ?? 0}${response.categoryTag ? `, тег ${response.categoryTag}` : ""}`;
    })
  };
}
