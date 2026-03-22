import {
  type AuthTokens,
  type DirectoryItem,
  type InvitationDraft,
  type InvitationLink,
  type QuestionDraft,
  type QuestionView,
  type ReportDelivery,
  type ReportFormat,
  type SessionAnalytics,
  type SessionRecord,
  type SessionResponse,
  type ShareLinkConfig,
  type ShareLinkDraft,
  type StartSessionInput,
  type StartSessionResult,
  type SubmitAnswerInput,
  type SubmitAnswerResult,
  type SurveyDraft,
  type SurveyRecord,
  type SurveySummary,
  type UserProfile
} from "./types";
import { buildQuestionTemplate, buildDefaultSettings, createId, encodeSetup, ensureQuestionOrder, safeParseJson } from "./utils";

type MockUser = UserProfile & {
  password: string;
};

type MockInvitation = InvitationDraft & {
  token: string;
  invitationUrl: string;
  createdAt: string;
  used: boolean;
  registeredAt?: string;
};

type MockDatabase = {
  users: MockUser[];
  invitations: MockInvitation[];
  surveys: SurveyRecord[];
  sessions: SessionRecord[];
};

const MOCK_DB_KEY = "profdnk.mock-db.v1";
const TOKEN_TTL_SECONDS = 900;

function nowIso() {
  return new Date().toISOString();
}

function todayPlus(days: number) {
  const date = new Date();
  date.setDate(date.getDate() + days);
  return date.toISOString();
}

function clone<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function getOrigin() {
  const configured = (import.meta.env.VITE_PUBLIC_APP_URL as string | undefined) ?? "https://hack.benzo.cloud";

  if (typeof window === "undefined") {
    return configured;
  }

  try {
    return new URL(configured, window.location.origin).toString().replace(/\/$/, "");
  } catch {
    return configured;
  }
}

function buildTokens(user: UserProfile): AuthTokens {
  return {
    accessToken: `mock-access:${user.id}:${user.role}`,
    refreshToken: `mock-refresh:${user.id}:${user.role}`,
    expiresIn: TOKEN_TTL_SECONDS,
    role: user.role
  };
}

function toUserProfile(user: MockUser): UserProfile {
  const { password: _password, ...profile } = user;
  return profile;
}

function readStorage() {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage.getItem(MOCK_DB_KEY);
}

function writeStorage(db: MockDatabase) {
  if (typeof window === "undefined") {
    return;
  }

  window.localStorage.setItem(MOCK_DB_KEY, JSON.stringify(db));
}

function parseToken(authorizationOrToken: string) {
  const raw = authorizationOrToken.startsWith("Bearer ")
    ? authorizationOrToken.replace("Bearer ", "")
    : authorizationOrToken;
  const [, userId] = raw.split(":");
  return userId;
}

function questionToView(question: QuestionDraft): QuestionView {
  return {
    questionId: question.id,
    type: question.type,
    text: question.text,
    helperText: question.helperText,
    answers: question.answers.map((answer) => ({
      answerId: answer.id,
      text: answer.text
    }))
  };
}

function refreshSurveyStats(db: MockDatabase, surveyId: string) {
  const completions = db.sessions.filter((session) => session.surveyId === surveyId && session.status === "completed").length;
  const survey = db.surveys.find((item) => item.surveyId === surveyId);

  if (survey) {
    survey.completionsCount = completions;
  }
}

function buildSurveyLink(survey: SurveyRecord, config: ShareLinkConfig) {
  const setup = encodeSetup({
    shareLinkId: config.id,
    title: config.title,
    description: config.description,
    intro: config.intro,
    ownerId: config.ownerId,
    fields: config.extraFields
  });

  return `${getOrigin()}/tests/${survey.surveyId}/start?setup=${encodeURIComponent(setup)}`;
}

function buildSampleSurvey(psychologistId: string): SurveyRecord {
  const q1 = buildQuestionTemplate("single_choice", 1);
  q1.id = "6c0d0893-1a30-4b23-b908-7d0f33750a01";
  q1.text = "Какой тип задач дает вам больше энергии?";
  q1.answers = [
    { id: "6c0d0893-1a30-4b23-b908-7d0f33750b01", text: "Коммуникация и сопровождение людей", weight: 3, categoryTag: "people" },
    { id: "6c0d0893-1a30-4b23-b908-7d0f33750b02", text: "Анализ, схемы и исследование", weight: 4, categoryTag: "analysis" }
  ];

  const q2 = buildQuestionTemplate("multiple_choice", 2);
  q2.id = "6c0d0893-1a30-4b23-b908-7d0f33750a02";
  q2.text = "Какие форматы вам подходят сразу несколько?";
  q2.answers = [
    { id: "6c0d0893-1a30-4b23-b908-7d0f33750b03", text: "Вести интервью", weight: 2, categoryTag: "people" },
    { id: "6c0d0893-1a30-4b23-b908-7d0f33750b04", text: "Собирать и проверять гипотезы", weight: 4, categoryTag: "analysis" },
    { id: "6c0d0893-1a30-4b23-b908-7d0f33750b05", text: "Проектировать маршрут клиента", weight: 3, categoryTag: "systems" }
  ];

  const q3 = buildQuestionTemplate("scale", 3);
  q3.id = "6c0d0893-1a30-4b23-b908-7d0f33750a03";
  q3.text = "Насколько вам близка работа в неопределенности?";

  const q4 = buildQuestionTemplate("text", 4);
  q4.id = "6c0d0893-1a30-4b23-b908-7d0f33750a04";
  q4.text = "Опишите, что вы хотите лучше понять о себе после теста.";

  q1.logicRules = [
    {
      answerId: q1.answers[0].id,
      action: "linear"
    },
    {
      answerId: q1.answers[1].id,
      action: "jump",
      nextQuestionId: q3.id
    }
  ];

  const survey: SurveyRecord = {
    surveyId: "6c0d0893-1a30-4b23-b908-7d0f33750001",
    psychologistId,
    title: "Профориентационный спринт",
    description: "Демо-методика для проверки сценариев кабинета, ссылок и отчетов.",
    status: "active",
    completionsCount: 0,
    createdAt: todayPlus(-5),
    settings: {
      ...buildDefaultSettings(),
      startForm: {
        ...buildDefaultSettings().startForm,
        intro: "Ответьте на несколько вопросов. Итоговый отчет придет на email сразу после завершения сессии."
      }
    },
    questions: [q1, q2, q3, q4],
    shareLinks: []
  };

  const defaultLink: ShareLinkConfig = {
    id: "c781e7b1-7b08-4d4a-a0ff-f7f4d745f001",
    surveyId: survey.surveyId,
    title: "Основная ссылка на диагностику",
    description: "Базовый сценарий прохождения теста",
    intro: "Перед стартом оставьте контакты, чтобы система отправила персональный отчет.",
    ownerId: psychologistId,
    extraFields: [
      {
        key: "grade",
        label: "Класс / курс",
        required: false,
        kind: "text",
        placeholder: "11 класс"
      }
    ],
    allowSelfReport: true,
    publicUrl: "",
    createdAt: todayPlus(-4)
  };

  defaultLink.publicUrl = buildSurveyLink(survey, defaultLink);
  survey.shareLinks = [defaultLink];

  return survey;
}

function buildSeed(): MockDatabase {
  const admin: MockUser = {
    id: "fb517330-3f02-4b8f-b753-278532b5f001",
    email: "admin@profdnk.local",
    fullName: "Системный администратор",
    phone: "+7 999 555-11-00",
    role: "admin",
    status: "active",
    photoUrl: "",
    about: "Управляет доступами и приглашениями.",
    accessUntil: todayPlus(365),
    password: "admin12345"
  };

  const psychologist: MockUser = {
    id: "fb517330-3f02-4b8f-b753-278532b5f002",
    email: "psycho@profdnk.local",
    fullName: "Анна Смирнова",
    phone: "+7 999 555-22-00",
    role: "psychologist",
    status: "active",
    photoUrl: "",
    about: "Профориентолог и карьерный консультант.",
    accessUntil: todayPlus(200),
    password: "psych12345"
  };

  const survey = buildSampleSurvey(psychologist.id);

  const sessions: SessionRecord[] = [
    {
      sessionId: "ab517330-3f02-4b8f-b753-278532b5f101",
      surveyId: survey.surveyId,
      shareLinkId: survey.shareLinks[0].id,
      clientMetadata: {
        fullName: "Иван Иванов",
        email: "ivan@example.com",
        grade: "11 класс"
      },
      currentQuestionId: "",
      currentIndex: 4,
      status: "completed",
      startedAt: todayPlus(-3),
      finishedAt: todayPlus(-3),
      responses: [
        {
          questionId: survey.questions[0].id,
          questionText: survey.questions[0].text,
          questionType: survey.questions[0].type,
          answerId: survey.questions[0].answers[1].id,
          selectedWeight: survey.questions[0].answers[1].weight,
          categoryTag: survey.questions[0].answers[1].categoryTag
        },
        {
          questionId: survey.questions[2].id,
          questionText: survey.questions[2].text,
          questionType: survey.questions[2].type,
          answerId: survey.questions[2].answers[4].id,
          selectedWeight: survey.questions[2].answers[4].weight
        },
        {
          questionId: survey.questions[3].id,
          questionText: survey.questions[3].text,
          questionType: survey.questions[3].type,
          rawText: "Хочу понять, где я полезнее всего и где смогу быстрее расти."
        }
      ],
      reportDelivery: {
        status: "sent",
        email: "ivan@example.com",
        fileName: "client-report.docx",
        contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
      }
    },
    {
      sessionId: "ab517330-3f02-4b8f-b753-278532b5f102",
      surveyId: survey.surveyId,
      shareLinkId: survey.shareLinks[0].id,
      clientMetadata: {
        fullName: "Мария Петрова",
        email: "maria@example.com",
        grade: "10 класс"
      },
      currentQuestionId: "",
      currentIndex: 4,
      status: "completed",
      startedAt: todayPlus(-1),
      finishedAt: todayPlus(-1),
      responses: [
        {
          questionId: survey.questions[0].id,
          questionText: survey.questions[0].text,
          questionType: survey.questions[0].type,
          answerId: survey.questions[0].answers[0].id,
          selectedWeight: survey.questions[0].answers[0].weight,
          categoryTag: survey.questions[0].answers[0].categoryTag
        },
        {
          questionId: survey.questions[1].id,
          questionText: survey.questions[1].text,
          questionType: survey.questions[1].type,
          answerIds: [survey.questions[1].answers[0].id, survey.questions[1].answers[2].id],
          selectedWeight: survey.questions[1].answers[0].weight + survey.questions[1].answers[2].weight,
          categoryTag: "people, systems"
        },
        {
          questionId: survey.questions[2].id,
          questionText: survey.questions[2].text,
          questionType: survey.questions[2].type,
          answerId: survey.questions[2].answers[2].id,
          selectedWeight: survey.questions[2].answers[2].weight
        },
        {
          questionId: survey.questions[3].id,
          questionText: survey.questions[3].text,
          questionType: survey.questions[3].type,
          rawText: "Важно понять, где я чувствую уверенность и могу работать с людьми."
        }
      ],
      reportDelivery: {
        status: "sent",
        email: "maria@example.com",
        fileName: "client-report.docx",
        contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
      }
    }
  ];

  const db: MockDatabase = {
    users: [admin, psychologist],
    invitations: [],
    surveys: [survey],
    sessions
  };

  refreshSurveyStats(db, survey.surveyId);
  return db;
}

function getDb() {
  const storage = readStorage();

  if (!storage) {
    const seed = buildSeed();
    writeStorage(seed);
    return seed;
  }

  const parsed = safeParseJson<MockDatabase | null>(storage, null);
  if (parsed) {
    return parsed;
  }

  const seed = buildSeed();
  writeStorage(seed);
  return seed;
}

function saveDb(db: MockDatabase) {
  writeStorage(db);
  return clone(db);
}

function getUserFromAuthorization(authorization: string) {
  const db = getDb();
  const userId = parseToken(authorization);
  const user = db.users.find((candidate) => candidate.id === userId);

  if (!user) {
    throw new Error("Пользователь не найден. Повторите вход.");
  }

  return user;
}

function ensureAdmin(authorization: string) {
  const user = getUserFromAuthorization(authorization);

  if (user.role !== "admin") {
    throw new Error("Доступно только администратору");
  }

  return user;
}

function ensurePsychologist(authorization: string) {
  const user = getUserFromAuthorization(authorization);

  if (user.role !== "psychologist") {
    throw new Error("Доступно только психологу");
  }

  if (user.status !== "active") {
    throw new Error("Аккаунт психолога неактивен");
  }

  return user;
}

function normalizeSurveyDraft(input: SurveyDraft): SurveyRecord {
  return {
    surveyId: input.surveyId ?? createId(),
    psychologistId: input.psychologistId,
    title: input.title,
    description: input.description,
    status: "active",
    completionsCount: 0,
    createdAt: nowIso(),
    settings: {
      ...input.settings,
      startForm: {
        ...input.settings.startForm,
        fields: input.settings.startForm.fields.length > 0 ? input.settings.startForm.fields : buildDefaultSettings().startForm.fields
      }
    },
    questions: ensureQuestionOrder(input.questions),
    shareLinks: []
  };
}

function findSurveyOrThrow(db: MockDatabase, surveyId: string) {
  const survey = db.surveys.find((candidate) => candidate.surveyId === surveyId);

  if (!survey) {
    throw new Error("Тест не найден");
  }

  return survey;
}

function findSessionOrThrow(db: MockDatabase, sessionId: string) {
  const session = db.sessions.find((candidate) => candidate.sessionId === sessionId);

  if (!session) {
    throw new Error("Сессия не найдена");
  }

  return session;
}

function findQuestionOrThrow(survey: SurveyRecord, questionId: string) {
  const question = survey.questions.find((candidate) => candidate.id === questionId);

  if (!question) {
    throw new Error("Вопрос не найден");
  }

  return question;
}

function linearNextQuestion(survey: SurveyRecord, questionId: string) {
  const currentIndex = survey.questions.findIndex((question) => question.id === questionId);
  return survey.questions[currentIndex + 1];
}

function resolveNextQuestionId(
  survey: SurveyRecord,
  question: QuestionDraft,
  payload: SubmitAnswerInput
) {
  const selectedAnswerIds = payload.answerIds && payload.answerIds.length > 0
    ? payload.answerIds
    : payload.answerId
      ? [payload.answerId]
      : [];

  const matchedRules = question.logicRules.filter((rule) => selectedAnswerIds.includes(rule.answerId));
  const finishRule = matchedRules.find((rule) => rule.action === "finish");
  if (finishRule) {
    return "";
  }

  const jumpRule = matchedRules.find((rule) => rule.action === "jump" && rule.nextQuestionId);
  if (jumpRule?.nextQuestionId) {
    return jumpRule.nextQuestionId;
  }

  return linearNextQuestion(survey, question.id)?.id ?? "";
}

function buildSessionResponse(question: QuestionDraft, payload: SubmitAnswerInput): SessionResponse {
  if (payload.rawText) {
    return {
      questionId: question.id,
      questionText: question.text,
      questionType: question.type,
      rawText: payload.rawText
    };
  }

  if (payload.answerIds && payload.answerIds.length > 0) {
    const answers = question.answers.filter((answer) => payload.answerIds?.includes(answer.id));
    return {
      questionId: question.id,
      questionText: question.text,
      questionType: question.type,
      answerIds: answers.map((answer) => answer.id),
      selectedWeight: answers.reduce((sum, answer) => sum + answer.weight, 0),
      categoryTag: answers.map((answer) => answer.categoryTag).filter(Boolean).join(", ")
    };
  }

  const selected = question.answers.find((answer) => answer.id === payload.answerId);

  return {
    questionId: question.id,
    questionText: question.text,
    questionType: question.type,
    answerId: selected?.id,
    selectedWeight: selected?.weight,
    categoryTag: selected?.categoryTag
  };
}

export const mockBackend = {
  reset() {
    saveDb(buildSeed());
  },

  async login(email: string, password: string) {
    const db = getDb();
    const user = db.users.find((candidate) => candidate.email.toLowerCase() === email.trim().toLowerCase());

    if (!user || user.password !== password) {
      throw new Error("Неверный email или пароль");
    }

    if (user.status === "blocked") {
      throw new Error("Аккаунт заблокирован администратором");
    }

    if (user.status === "inactive") {
      throw new Error("Срок доступа истек");
    }

    return clone(buildTokens(user));
  },

  async refreshToken(refreshToken: string) {
    const db = getDb();
    const userId = parseToken(refreshToken);
    const user = db.users.find((candidate) => candidate.id === userId);

    if (!user) {
      throw new Error("Сессия истекла");
    }

    return clone(buildTokens(user));
  },

  async register(token: string, password: string) {
    const db = getDb();
    const invitation = db.invitations.find((candidate) => candidate.token === token);

    if (!invitation) {
      throw new Error("Приглашение не найдено");
    }

    if (invitation.used) {
      throw new Error("Ссылка уже использована");
    }

    if (new Date(invitation.expiresAt) < new Date()) {
      throw new Error("Срок действия приглашения истек");
    }

    const user: MockUser = {
      id: createId(),
      email: invitation.email,
      fullName: invitation.fullName,
      phone: invitation.phone,
      role: "psychologist",
      status: "active",
      photoUrl: "",
      about: "Новый специалист, приглашенный администратором.",
      accessUntil: invitation.accessUntil,
      password
    };

    invitation.used = true;
    invitation.registeredAt = nowIso();
    db.users.push(user);
    saveDb(db);

    return clone(buildTokens(user));
  },

  async getProfile(authorization: string) {
    return clone(toUserProfile(getUserFromAuthorization(authorization)));
  },

  async updateProfile(authorization: string, input: Partial<Pick<UserProfile, "photoUrl" | "about">>) {
    const db = getDb();
    const userId = parseToken(authorization);
    const user = db.users.find((candidate) => candidate.id === userId);

    if (!user) {
      throw new Error("Профиль не найден");
    }

    user.photoUrl = input.photoUrl ?? user.photoUrl;
    user.about = input.about ?? user.about;
    saveDb(db);

    return clone(toUserProfile(user));
  },

  async listPsychologists(authorization: string) {
    ensureAdmin(authorization);
    const db = getDb();

    const registered: DirectoryItem[] = db.users
      .filter((user) => user.role === "psychologist")
      .map((user) => ({
        id: user.id,
        fullName: user.fullName,
        phone: user.phone,
        email: user.email,
        role: user.role,
        status: user.status,
        accessUntil: user.accessUntil
      }));

    const pending: DirectoryItem[] = db.invitations
      .filter((item) => !item.used)
      .map((invitation) => ({
        fullName: invitation.fullName,
        phone: invitation.phone,
        email: invitation.email,
        role: invitation.role,
        status: "pending",
        accessUntil: invitation.accessUntil,
        expiresAt: invitation.expiresAt,
        invitationUrl: invitation.invitationUrl,
        invitationToken: invitation.token
      }));

    return [...registered, ...pending].sort((left, right) => left.fullName.localeCompare(right.fullName, "ru"));
  },

  async createInvitation(authorization: string, draft: InvitationDraft): Promise<InvitationLink> {
    ensureAdmin(authorization);
    const db = getDb();

    db.invitations = db.invitations.filter(
      (item) => !(item.email.toLowerCase() === draft.email.toLowerCase() && !item.used)
    );

    const token = createId();
    const invitation: MockInvitation = {
      ...draft,
      token,
      createdAt: nowIso(),
      used: false,
      invitationUrl: `${getOrigin()}/invitations/${token}`
    };

    db.invitations.push(invitation);
    saveDb(db);

    return {
      invitationToken: invitation.token,
      invitationUrl: invitation.invitationUrl
    };
  },

  async blockUser(authorization: string, email: string) {
    ensureAdmin(authorization);
    const db = getDb();
    const user = db.users.find((candidate) => candidate.email.toLowerCase() === email.trim().toLowerCase());

    if (!user) {
      throw new Error("Психолог с таким email не найден");
    }

    user.status = "blocked";
    saveDb(db);
  },

  async unblockUser(authorization: string, email: string) {
    ensureAdmin(authorization);
    const db = getDb();
    const user = db.users.find((candidate) => candidate.email.toLowerCase() === email.trim().toLowerCase());

    if (!user) {
      throw new Error("Психолог с таким email не найден");
    }

    user.status = "active";
    saveDb(db);
  },

  async listSurveys(authorization: string, psychologistId: string): Promise<SurveySummary[]> {
    const user = getUserFromAuthorization(authorization);
    if (user.role === "psychologist" && user.id !== psychologistId) {
      throw new Error("Нельзя смотреть тесты другого психолога");
    }

    const db = getDb();
    return db.surveys
      .filter((survey) => survey.psychologistId === psychologistId)
      .map((survey) => ({
        surveyId: survey.surveyId,
        title: survey.title,
        completionsCount: survey.completionsCount,
        status: survey.status
      }));
  },

  async getSurvey(authorization: string, surveyId: string) {
    ensurePsychologist(authorization);
    const db = getDb();
    return clone(findSurveyOrThrow(db, surveyId));
  },

  async createSurvey(authorization: string, draft: SurveyDraft) {
    const user = ensurePsychologist(authorization);
    const db = getDb();
    const survey = normalizeSurveyDraft({
      ...draft,
      psychologistId: user.id
    });

    db.surveys.push(survey);
    saveDb(db);
    return { surveyId: survey.surveyId };
  },

  async updateSurvey(authorization: string, surveyId: string, draft: SurveyDraft) {
    const user = ensurePsychologist(authorization);
    const db = getDb();
    const survey = findSurveyOrThrow(db, surveyId);

    if (survey.psychologistId !== user.id) {
      throw new Error("Нельзя редактировать чужой тест");
    }

    survey.title = draft.title;
    survey.description = draft.description;
    survey.settings = draft.settings;
    survey.questions = ensureQuestionOrder(draft.questions);
    saveDb(db);

    return clone(survey);
  },

  async annulSurvey(authorization: string, surveyId: string) {
    const user = ensurePsychologist(authorization);
    const db = getDb();
    const survey = findSurveyOrThrow(db, surveyId);

    if (survey.psychologistId !== user.id) {
      throw new Error("Нельзя аннулировать чужой тест");
    }

    survey.status = "annulled";
    saveDb(db);
    return clone(survey);
  },

  async createShareLink(authorization: string, surveyId: string, draft: ShareLinkDraft) {
    const user = ensurePsychologist(authorization);
    const db = getDb();
    const survey = findSurveyOrThrow(db, surveyId);

    if (survey.psychologistId !== user.id) {
      throw new Error("Нельзя управлять чужим тестом");
    }

    const config: ShareLinkConfig = {
      id: createId(),
      surveyId,
      title: draft.title,
      description: draft.description,
      intro: draft.intro,
      ownerId: user.id,
      extraFields: draft.extraFields,
      allowSelfReport: draft.allowSelfReport,
      publicUrl: "",
      createdAt: nowIso()
    };

    config.publicUrl = buildSurveyLink(survey, config);
    survey.shareLinks.push(config);
    saveDb(db);

    return clone(config);
  },

  async listSurveySessions(authorization: string, surveyId: string) {
    const user = ensurePsychologist(authorization);
    const db = getDb();
    const survey = findSurveyOrThrow(db, surveyId);

    if (survey.psychologistId !== user.id) {
      throw new Error("Нельзя смотреть сессии чужого теста");
    }

    return clone(db.sessions.filter((session) => session.surveyId === surveyId && session.status === "completed"));
  },

  async startSession(input: StartSessionInput): Promise<StartSessionResult> {
    const db = getDb();
    const survey = findSurveyOrThrow(db, input.surveyId);

    if (survey.status === "annulled") {
      throw new Error("Тест аннулирован и недоступен");
    }

    const firstQuestion = survey.questions[0];
    if (!firstQuestion) {
      throw new Error("В тесте нет вопросов");
    }

    const session: SessionRecord = {
      sessionId: createId(),
      surveyId: survey.surveyId,
      shareLinkId: input.shareLinkId,
      clientMetadata: input.clientMetadata,
      currentQuestionId: firstQuestion.id,
      currentIndex: 0,
      status: "active",
      startedAt: nowIso(),
      responses: []
    };

    db.sessions.push(session);
    saveDb(db);

    return {
      sessionId: session.sessionId,
      firstQuestion: questionToView(firstQuestion)
    };
  },

  async getCurrentQuestion(sessionId: string) {
    const db = getDb();
    const session = findSessionOrThrow(db, sessionId);
    const survey = findSurveyOrThrow(db, session.surveyId);

    if (session.status !== "active") {
      throw new Error("Сессия уже закрыта");
    }

    const question = findQuestionOrThrow(survey, session.currentQuestionId);
    return questionToView(question);
  },

  async submitAnswer(input: SubmitAnswerInput): Promise<SubmitAnswerResult> {
    const db = getDb();
    const session = findSessionOrThrow(db, input.sessionId);
    const survey = findSurveyOrThrow(db, session.surveyId);

    if (session.status !== "active") {
      throw new Error("Сессию больше нельзя продолжить");
    }

    const question = findQuestionOrThrow(survey, input.questionId);
    const response = buildSessionResponse(question, input);
    const existingIndex = session.responses.findIndex((item) => item.questionId === question.id);

    if (existingIndex >= 0) {
      session.responses[existingIndex] = response;
    } else {
      session.responses.push(response);
    }

    const nextQuestionId = resolveNextQuestionId(survey, question, input);

    if (!nextQuestionId) {
      session.status = "completed";
      session.finishedAt = nowIso();
      session.currentIndex = survey.questions.length;
      session.currentQuestionId = "";
      session.reportDelivery = {
        status: "sent",
        email: String(session.clientMetadata.email ?? ""),
        fileName: "client-report.docx",
        contentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
      };

      refreshSurveyStats(db, survey.surveyId);
      saveDb(db);

      return {
        nextQuestionId: "",
        isFinished: true,
        reportDelivery: clone(session.reportDelivery)
      };
    }

    const nextQuestion = findQuestionOrThrow(survey, nextQuestionId);
    session.currentQuestionId = nextQuestion.id;
    session.currentIndex = survey.questions.findIndex((item) => item.id === nextQuestion.id);
    saveDb(db);

    return {
      nextQuestionId: nextQuestion.id,
      isFinished: false,
      nextQuestion: questionToView(nextQuestion)
    };
  },

  async getSessionAnalytics(authorization: string, sessionId: string): Promise<SessionAnalytics> {
    ensurePsychologist(authorization);
    const db = getDb();
    const session = findSessionOrThrow(db, sessionId);
    return clone({
      surveyId: session.surveyId,
      sessionId: session.sessionId,
      clientMetadata: session.clientMetadata,
      responses: session.responses,
      startedAt: session.startedAt,
      finishedAt: session.finishedAt
    });
  },

  async sendSessionReport(authorization: string, sessionId: string, reportFormat: ReportFormat): Promise<ReportDelivery> {
    const user = ensurePsychologist(authorization);
    const db = getDb();
    const session = findSessionOrThrow(db, sessionId);
    const survey = findSurveyOrThrow(db, session.surveyId);

    if (survey.psychologistId !== user.id) {
      throw new Error("Нельзя работать с чужими отчетами");
    }

    const email = reportFormat.startsWith("psycho")
      ? user.email
      : String(session.clientMetadata.email ?? "");

    const report: ReportDelivery = {
      status: "sent",
      email,
      fileName: reportFormat.startsWith("psycho") ? "psychologist-report.docx" : "client-report.docx",
      contentType: reportFormat.endsWith("html")
        ? "text/html"
        : "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
    };

    session.reportDelivery = report;
    saveDb(db);

    return clone(report);
  }
};
