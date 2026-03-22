import { mockBackend } from "./mock-backend";
import type {
  ApiMode,
  AuthTokens,
  DirectoryItem,
  InvitationDraft,
  InvitationLink,
  QuestionView,
  QuestionDraft,
  ReportDelivery,
  ReportFormat,
  SessionAnalytics,
  SessionRecord,
  ShareLinkConfig,
  ShareLinkDraft,
  StartSessionInput,
  StartSessionResult,
  SubmitAnswerInput,
  SubmitAnswerResult,
  SurveyDraft,
  SurveyRecord,
  SurveySummary,
  UserProfile
} from "./types";
import { createId, encodeSetup, safeParseJson } from "./utils";

type WorkspaceState = {
  directory: DirectoryItem[];
  shareLinksBySurvey: Record<string, ShareLinkConfig[]>;
  draftSurveys: Record<string, SurveyRecord>;
  annulledSurveyIds: string[];
};

function resolveApiBaseUrl() {
  const configured = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? "http://localhost:8080";

  if (typeof window === "undefined") {
    return configured;
  }

  try {
    const url = new URL(configured, window.location.origin);
    const browserHost = window.location.hostname;
    const isConfiguredLoopback = ["localhost", "127.0.0.1", "0.0.0.0"].includes(url.hostname);
    const isBrowserLoopback = ["localhost", "127.0.0.1", "0.0.0.0"].includes(browserHost);

    if (isConfiguredLoopback && !isBrowserLoopback) {
      url.hostname = browserHost;
    }

    return url.toString().replace(/\/$/, "");
  } catch {
    return configured;
  }
}

const API_BASE_URL = resolveApiBaseUrl();
const API_MODE = ((import.meta.env.VITE_API_MODE as ApiMode | undefined) ?? "mock") as ApiMode;
const WORKSPACE_KEY = "profdnk.workspace.v1";

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function readWorkspace(): WorkspaceState {
  if (typeof window === "undefined") {
    return {
      directory: [],
      shareLinksBySurvey: {},
      draftSurveys: {},
      annulledSurveyIds: []
    };
  }

  const raw = window.localStorage.getItem(WORKSPACE_KEY);
  return safeParseJson<WorkspaceState>(raw, {
    directory: [],
    shareLinksBySurvey: {},
    draftSurveys: {},
    annulledSurveyIds: []
  });
}

function writeWorkspace(state: WorkspaceState) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(WORKSPACE_KEY, JSON.stringify(state));
}

function upsertDirectoryItem(item: DirectoryItem) {
  const workspace = readWorkspace();
  const existing = workspace.directory.findIndex((candidate) => candidate.email.toLowerCase() === item.email.toLowerCase());

  if (existing >= 0) {
    workspace.directory[existing] = {
      ...workspace.directory[existing],
      ...item
    };
  } else {
    workspace.directory.unshift(item);
  }

  writeWorkspace(workspace);
}

function storeDraftSurvey(survey: SurveyRecord) {
  const workspace = readWorkspace();
  workspace.draftSurveys[survey.surveyId] = clone(survey);
  writeWorkspace(workspace);
}

function readStoredSurvey(surveyId: string) {
  return readWorkspace().draftSurveys[surveyId];
}

function listStoredShareLinks(surveyId: string) {
  return readWorkspace().shareLinksBySurvey[surveyId] ?? [];
}

function saveShareLink(surveyId: string, link: ShareLinkConfig) {
  const workspace = readWorkspace();
  const links = workspace.shareLinksBySurvey[surveyId] ?? [];
  workspace.shareLinksBySurvey[surveyId] = [link, ...links];
  writeWorkspace(workspace);
}

function markAnnulledSurvey(surveyId: string) {
  const workspace = readWorkspace();
  if (!workspace.annulledSurveyIds.includes(surveyId)) {
    workspace.annulledSurveyIds.push(surveyId);
    writeWorkspace(workspace);
  }
}

async function parseResponse<T>(response: Response): Promise<T> {
  if (response.ok) {
    if (response.status === 204) {
      return undefined as T;
    }

    return (await response.json()) as T;
  }

  try {
    const payload = (await response.json()) as { error?: { message?: string } };
    throw new Error(payload.error?.message ?? "Сервис вернул ошибку");
  } catch (error) {
    if (error instanceof Error) {
      throw error;
    }

    throw new Error("Сервис вернул ошибку");
  }
}

async function request<T>(path: string, init?: RequestInit & { accessToken?: string }): Promise<T> {
  const headers = new Headers(init?.headers);
  headers.set("Accept", "application/json");

  if (init?.body) {
    headers.set("Content-Type", "application/json");
  }

  if (init?.accessToken) {
    headers.set("Authorization", `Bearer ${init.accessToken}`);
  }

  let response: Response;
  try {
    response = await fetch(`${API_BASE_URL}${path}`, {
      ...init,
      headers
    });
  } catch (error) {
    if (error instanceof TypeError) {
      throw new Error(`Не удалось подключиться к BFF по адресу ${API_BASE_URL}. Проверьте, что сервис запущен и доступен из браузера.`);
    }

    throw error;
  }

  return parseResponse<T>(response);
}

function mapSurveyToPayload(draft: SurveyDraft) {
  return {
    psychologistId: draft.psychologistId,
    title: draft.title,
    description: draft.description,
    settings: {
      limits: {
        time_limit_sec: draft.settings.limits.timeLimitSec
      },
      start_form: {
        fields: draft.settings.startForm.fields.map((field) => field.key),
        intro: draft.settings.startForm.intro,
        completion_title: draft.settings.startForm.completionTitle,
        completion_body: draft.settings.startForm.completionBody
      }
    },
    questions: draft.questions.map((question: QuestionDraft) => ({
      orderNum: question.orderNum,
      type: question.type,
      text: question.text,
      logicRules: {
        rules: question.logicRules,
        helperText: question.helperText
      },
      answers: question.answers.map((answer) => ({
        id: answer.id,
        text: answer.text,
        weight: answer.weight,
        categoryTag: answer.categoryTag
      }))
    }))
  };
}

function buildLocalShareLink(surveyId: string, survey: SurveyRecord, draft: ShareLinkDraft): ShareLinkConfig {
  const link: ShareLinkConfig = {
    id: createId(),
    surveyId,
    title: draft.title,
    description: draft.description,
    intro: draft.intro,
    ownerId: survey.psychologistId,
    extraFields: draft.extraFields,
    allowSelfReport: draft.allowSelfReport,
    publicUrl: "",
    createdAt: new Date().toISOString()
  };

  const setup = encodeSetup({
    shareLinkId: link.id,
    title: link.title,
    description: link.description,
    intro: link.intro,
    ownerId: link.ownerId,
    fields: link.extraFields
  });
  link.publicUrl = `${window.location.origin}/tests/${surveyId}/start?setup=${encodeURIComponent(setup)}`;
  return link;
}

export const api = {
  mode: API_MODE,

  async login(email: string, password: string) {
    if (API_MODE === "mock") {
      return mockBackend.login(email, password);
    }

    return request<AuthTokens>("/public/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password })
    });
  },

  async refreshToken(refreshToken: string) {
    if (API_MODE === "mock") {
      return mockBackend.refreshToken(refreshToken);
    }

    return request<AuthTokens>("/public/v1/auth/refresh", {
      method: "POST",
      body: JSON.stringify({ refreshToken })
    });
  },

  async register(token: string, password: string) {
    if (API_MODE === "mock") {
      return mockBackend.register(token, password);
    }

    return request<AuthTokens>("/public/v1/auth/register", {
      method: "POST",
      body: JSON.stringify({ token, password })
    });
  },

  async getProfile(accessToken: string): Promise<UserProfile> {
    if (API_MODE === "mock") {
      return mockBackend.getProfile(`Bearer ${accessToken}`);
    }

    return request<UserProfile>("/api/v1/auth/profile", {
      accessToken
    });
  },

  async updateProfile(accessToken: string, input: Partial<Pick<UserProfile, "photoUrl" | "about">>) {
    if (API_MODE === "mock") {
      return mockBackend.updateProfile(`Bearer ${accessToken}`, input);
    }

    return request<UserProfile>("/api/v1/auth/profile", {
      method: "PATCH",
      accessToken,
      body: JSON.stringify(input)
    });
  },

  async listPsychologists(accessToken: string): Promise<DirectoryItem[]> {
    if (API_MODE === "mock") {
      return mockBackend.listPsychologists(`Bearer ${accessToken}`);
    }

    return readWorkspace().directory;
  },

  async createInvitation(accessToken: string, draft: InvitationDraft): Promise<InvitationLink> {
    if (API_MODE === "mock") {
      return mockBackend.createInvitation(`Bearer ${accessToken}`, draft);
    }

    const response = await request<InvitationLink>("/api/v1/auth/invitations", {
      method: "POST",
      accessToken,
      body: JSON.stringify(draft)
    });

    upsertDirectoryItem({
      fullName: draft.fullName,
      phone: draft.phone,
      email: draft.email,
      role: draft.role,
      status: "pending",
      accessUntil: draft.accessUntil,
      expiresAt: draft.expiresAt,
      invitationUrl: response.invitationUrl,
      invitationToken: response.invitationToken
    });

    return response;
  },

  async blockUser(accessToken: string, email: string) {
    if (API_MODE === "mock") {
      return mockBackend.blockUser(`Bearer ${accessToken}`, email);
    }

    await request("/api/v1/auth/users/block", {
      method: "POST",
      accessToken,
      body: JSON.stringify({ email })
    });

    upsertDirectoryItem({
      fullName: email,
      phone: "",
      email,
      role: "psychologist",
      status: "blocked",
      accessUntil: ""
    });
  },

  async unblockUser(accessToken: string, email: string) {
    if (API_MODE === "mock") {
      return mockBackend.unblockUser(`Bearer ${accessToken}`, email);
    }

    await request("/api/v1/auth/users/unblock", {
      method: "POST",
      accessToken,
      body: JSON.stringify({ email })
    });

    upsertDirectoryItem({
      fullName: email,
      phone: "",
      email,
      role: "psychologist",
      status: "active",
      accessUntil: ""
    });
  },

  async listSurveys(accessToken: string, psychologistId: string): Promise<SurveySummary[]> {
    if (API_MODE === "mock") {
      return mockBackend.listSurveys(`Bearer ${accessToken}`, psychologistId);
    }

    const response = await request<{ surveys: SurveySummary[] }>(`/api/v1/surveys?psychologistId=${psychologistId}`, {
      accessToken
    });

    const annulled = new Set(readWorkspace().annulledSurveyIds);
    return response.surveys.map((survey) => ({
      ...survey,
      status: annulled.has(survey.surveyId) ? "annulled" : survey.status ?? "active"
    }));
  },

  async getSurvey(accessToken: string, surveyId: string) {
    if (API_MODE === "mock") {
      return mockBackend.getSurvey(`Bearer ${accessToken}`, surveyId);
    }

    const cached = readStoredSurvey(surveyId);
    if (cached) {
      return cached;
    }

    return request<SurveyRecord>(`/api/v1/surveys/${surveyId}`, {
      accessToken
    });
  },

  async createSurvey(accessToken: string, draft: SurveyDraft) {
    if (API_MODE === "mock") {
      return mockBackend.createSurvey(`Bearer ${accessToken}`, draft);
    }

    const response = await request<{ surveyId: string }>("/api/v1/surveys", {
      method: "POST",
      accessToken,
      body: JSON.stringify(mapSurveyToPayload(draft))
    });

    storeDraftSurvey({
      surveyId: response.surveyId,
      psychologistId: draft.psychologistId,
      title: draft.title,
      description: draft.description,
      status: "active",
      completionsCount: 0,
      createdAt: new Date().toISOString(),
      settings: draft.settings,
      questions: draft.questions,
      shareLinks: []
    });

    return response;
  },

  async updateSurvey(accessToken: string, surveyId: string, draft: SurveyDraft) {
    if (API_MODE === "mock") {
      return mockBackend.updateSurvey(`Bearer ${accessToken}`, surveyId, draft);
    }

    const updatedSurvey: SurveyRecord = {
      surveyId,
      psychologistId: draft.psychologistId,
      title: draft.title,
      description: draft.description,
      status: "active",
      completionsCount: 0,
      createdAt: new Date().toISOString(),
      settings: draft.settings,
      questions: draft.questions,
      shareLinks: listStoredShareLinks(surveyId)
    };

    storeDraftSurvey(updatedSurvey);

    try {
      return await request<SurveyRecord>(`/api/v1/surveys/${surveyId}`, {
        method: "PATCH",
        accessToken,
        body: JSON.stringify(mapSurveyToPayload(draft))
      });
    } catch {
      return updatedSurvey;
    }
  },

  async annulSurvey(accessToken: string, surveyId: string) {
    if (API_MODE === "mock") {
      return mockBackend.annulSurvey(`Bearer ${accessToken}`, surveyId);
    }

    markAnnulledSurvey(surveyId);
    const survey = readStoredSurvey(surveyId);
    if (survey) {
      survey.status = "annulled";
      storeDraftSurvey(survey);
    }

    return survey;
  },

  async createShareLink(accessToken: string, surveyId: string, draft: ShareLinkDraft) {
    if (API_MODE === "mock") {
      return mockBackend.createShareLink(`Bearer ${accessToken}`, surveyId, draft);
    }

    const survey = readStoredSurvey(surveyId);
    if (!survey) {
      throw new Error("Чтобы создавать публичные ссылки в режиме BFF, сначала сохраните тест в локальном конструкторе.");
    }

    const link = buildLocalShareLink(surveyId, survey, draft);
    saveShareLink(surveyId, link);
    survey.shareLinks = [link, ...survey.shareLinks];
    storeDraftSurvey(survey);

    return link;
  },

  async listSurveySessions(accessToken: string, surveyId: string): Promise<SessionRecord[]> {
    if (API_MODE === "mock") {
      return mockBackend.listSurveySessions(`Bearer ${accessToken}`, surveyId);
    }

    try {
      return await request<SessionRecord[]>(`/api/v1/surveys/${surveyId}/sessions`, {
        accessToken
      });
    } catch {
      return [];
    }
  },

  async startSession(input: StartSessionInput): Promise<StartSessionResult> {
    if (API_MODE === "mock") {
      return mockBackend.startSession(input);
    }

    return request<StartSessionResult>("/public/v1/sessions", {
      method: "POST",
      body: JSON.stringify(input)
    });
  },

  async getCurrentQuestion(sessionId: string): Promise<QuestionView> {
    if (API_MODE === "mock") {
      return mockBackend.getCurrentQuestion(sessionId);
    }

    return request<QuestionView>(`/public/v1/sessions/${sessionId}/current-question`);
  },

  async submitAnswer(input: SubmitAnswerInput): Promise<SubmitAnswerResult> {
    if (API_MODE === "mock") {
      return mockBackend.submitAnswer(input);
    }

    return request<SubmitAnswerResult>(`/public/v1/sessions/${input.sessionId}/answers`, {
      method: "POST",
      body: JSON.stringify({
        questionId: input.questionId,
        answerId: input.answerId,
        answerIds: input.answerIds,
        rawText: input.rawText
      })
    });
  },

  async getSessionAnalytics(accessToken: string, sessionId: string): Promise<SessionAnalytics> {
    if (API_MODE === "mock") {
      return mockBackend.getSessionAnalytics(`Bearer ${accessToken}`, sessionId);
    }

    return request<SessionAnalytics>(`/api/v1/sessions/${sessionId}/analytics`, {
      accessToken
    });
  },

  async sendSessionReport(accessToken: string, sessionId: string, reportFormat: ReportFormat): Promise<ReportDelivery> {
    if (API_MODE === "mock") {
      return mockBackend.sendSessionReport(`Bearer ${accessToken}`, sessionId, reportFormat);
    }

    return request<ReportDelivery>(`/api/v1/sessions/${sessionId}/report/send`, {
      method: "POST",
      accessToken,
      body: JSON.stringify({ reportFormat })
    });
  }
};
