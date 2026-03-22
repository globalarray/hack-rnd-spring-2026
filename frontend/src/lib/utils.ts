import type { CandidateSetup, QuestionDraft, StartFormField, SurveySettings } from "./types";

export const DEFAULT_FIELDS: StartFormField[] = [
  {
    key: "fullName",
    label: "ФИО",
    required: true,
    kind: "text",
    placeholder: "Иванов Иван Иванович"
  },
  {
    key: "email",
    label: "Email",
    required: true,
    kind: "email",
    placeholder: "name@example.com"
  }
];

export function createId() {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }

  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, (char) => {
    const random = Math.random() * 16 | 0;
    const value = char === "x" ? random : (random & 0x3) | 0x8;
    return value.toString(16);
  });
}

export function safeParseJson<T>(value: string | null | undefined, fallback: T) {
  if (!value) {
    return fallback;
  }

  try {
    return JSON.parse(value) as T;
  } catch {
    return fallback;
  }
}

export function formatDate(value?: string) {
  if (!value) {
    return "—";
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "long",
    year: "numeric"
  }).format(new Date(value));
}

export function formatDateTime(value?: string) {
  if (!value) {
    return "—";
  }

  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit"
  }).format(new Date(value));
}

export function readErrorMessage(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }

  return "Не удалось выполнить запрос";
}

export function buildDefaultSettings(): SurveySettings {
  return {
    limits: {
      timeLimitSec: 1200
    },
    startForm: {
      intro: "Заполните короткую анкету. После этого мы откроем вопросы и отправим готовый отчет на ваш email.",
      completionTitle: "Отчет уже в пути",
      completionBody: "Спасибо за прохождение. Как только аналитика соберет документ, письмо уйдет на указанный email.",
      fields: DEFAULT_FIELDS
    }
  };
}

export function buildQuestionTemplate(type: QuestionDraft["type"], orderNum: number): QuestionDraft {
  if (type === "text") {
    return {
      id: createId(),
      orderNum,
      type,
      text: "Опишите своими словами, что вам особенно интересно в работе или учебе.",
      helperText: "Свободный ответ поможет психологу точнее интерпретировать результаты.",
      answers: [],
      logicRules: []
    };
  }

  if (type === "scale") {
    const answers = Array.from({ length: 5 }, (_, index) => ({
      id: createId(),
      text: String(index + 1),
      weight: index + 1
    }));

    return {
      id: createId(),
      orderNum,
      type,
      text: "Насколько вам интересно разбираться в новых системах и подходах?",
      helperText: "Шкала от 1 до 5, где 1 — совсем неинтересно, 5 — очень интересно.",
      answers,
      logicRules: []
    };
  }

  return {
    id: createId(),
    orderNum,
    type,
    text: "Какой формат деятельности вам ближе?",
    helperText: "Можно адаптировать под нужную методику.",
    answers: [
      {
        id: createId(),
        text: "Работа с людьми",
        weight: 1,
        categoryTag: "people"
      },
      {
        id: createId(),
        text: "Работа с данными",
        weight: 2,
        categoryTag: "analysis"
      }
    ],
    logicRules: []
  };
}

export function ensureQuestionOrder(questions: QuestionDraft[]) {
  return questions.map((question, index) => ({
    ...question,
    orderNum: index + 1
  }));
}

export function encodeSetup(setup: CandidateSetup) {
  const encoder = new TextEncoder();
  const bytes = encoder.encode(JSON.stringify(setup));
  const chars = Array.from(bytes, (byte) => String.fromCharCode(byte)).join("");
  return btoa(chars);
}

export function decodeSetup(value: string | null): CandidateSetup | null {
  if (!value) {
    return null;
  }

  try {
    const binary = atob(value);
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0));
    const decoder = new TextDecoder();
    return JSON.parse(decoder.decode(bytes)) as CandidateSetup;
  } catch {
    return null;
  }
}

export function normalizeFieldKey(label: string) {
  return label
    .trim()
    .toLowerCase()
    .replace(/[^a-zA-Z0-9а-яА-Я]+/g, "_")
    .replace(/^_+|_+$/g, "")
    .replace(/_+/g, "_");
}

export function surveyToStartFieldKeys(fields: StartFormField[]) {
  return fields.map((field) => field.key);
}

export function toQuestionOptions(questions: QuestionDraft[]) {
  return questions.map((question) => ({
    value: question.id,
    label: `${question.orderNum}. ${question.text || "Без текста"}`
  }));
}

export function cx(...values: Array<string | false | null | undefined>) {
  return values.filter(Boolean).join(" ");
}

export async function copyText(value: string) {
  if (typeof window !== "undefined" && typeof navigator !== "undefined" && navigator.clipboard && window.isSecureContext) {
    await navigator.clipboard.writeText(value);
    return;
  }

  if (typeof document === "undefined") {
    throw new Error("Буфер обмена недоступен в текущем окружении.");
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "true");
  textarea.style.position = "fixed";
  textarea.style.top = "0";
  textarea.style.left = "0";
  textarea.style.opacity = "0";

  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  textarea.setSelectionRange(0, textarea.value.length);

  const copied = document.execCommand("copy");
  document.body.removeChild(textarea);

  if (!copied) {
    throw new Error("Не удалось скопировать ссылку автоматически. Скопируйте ее вручную.");
  }
}
