export type ApiMode = "mock" | "bff" | "hybrid";
export type Role = "admin" | "psychologist";
export type AccountStatus = "active" | "inactive" | "blocked";
export type QuestionType = "single_choice" | "multiple_choice" | "scale" | "text";
export type LogicAction = "linear" | "jump" | "finish";
export type SurveyStatus = "active" | "annulled";
export type SessionStatus = "active" | "completed";
export type ReportFormat = "client_docx" | "client_html" | "psycho_docx" | "psycho_html";
export type ReportDeliveryStatus = "sent" | "failed";
export type StartFieldKind = "text" | "email" | "number" | "tel";

export type AuthTokens = {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  role: Role;
};

export type UserProfile = {
  id: string;
  email: string;
  fullName: string;
  phone: string;
  role: Role;
  status: AccountStatus;
  photoUrl: string;
  about: string;
  accessUntil: string;
};

export type InvitationDraft = {
  fullName: string;
  phone: string;
  email: string;
  role: Role;
  accessUntil: string;
  expiresAt: string;
};

export type InvitationLink = {
  invitationToken: string;
  invitationUrl: string;
};

export type DirectoryItem = {
  id?: string;
  fullName: string;
  phone: string;
  email: string;
  role: Role;
  status: AccountStatus | "pending";
  accessUntil: string;
  expiresAt?: string;
  invitationUrl?: string;
  invitationToken?: string;
  registeredAt?: string;
};

export type StartFormField = {
  key: string;
  label: string;
  required: boolean;
  kind: StartFieldKind;
  placeholder?: string;
};

export type SurveySettings = {
  limits: {
    timeLimitSec: number;
  };
  startForm: {
    intro: string;
    completionTitle: string;
    completionBody: string;
    fields: StartFormField[];
  };
};

export type AnswerDraft = {
  id: string;
  text: string;
  weight: number;
  categoryTag?: string;
};

export type LogicRuleDraft = {
  answerId: string;
  action: LogicAction;
  nextQuestionId?: string;
};

export type QuestionDraft = {
  id: string;
  orderNum: number;
  type: QuestionType;
  text: string;
  helperText: string;
  answers: AnswerDraft[];
  logicRules: LogicRuleDraft[];
};

export type ShareLinkConfig = {
  id: string;
  surveyId: string;
  title: string;
  description: string;
  intro: string;
  ownerId: string;
  extraFields: StartFormField[];
  allowSelfReport: boolean;
  publicUrl: string;
  createdAt: string;
};

export type ShareLinkDraft = {
  title: string;
  description: string;
  intro: string;
  extraFields: StartFormField[];
  allowSelfReport: boolean;
};

export type SurveyRecord = {
  surveyId: string;
  psychologistId: string;
  title: string;
  description: string;
  status: SurveyStatus;
  completionsCount: number;
  createdAt: string;
  settings: SurveySettings;
  questions: QuestionDraft[];
  shareLinks: ShareLinkConfig[];
};

export type SurveySummary = {
  surveyId: string;
  title: string;
  completionsCount: number;
  status: SurveyStatus;
};

export type SurveyDraft = {
  surveyId?: string;
  psychologistId: string;
  title: string;
  description: string;
  settings: SurveySettings;
  questions: QuestionDraft[];
};

export type CandidateAnswer = {
  answerId: string;
  text: string;
};

export type QuestionView = {
  questionId: string;
  type: QuestionType;
  text: string;
  helperText?: string;
  answers: CandidateAnswer[];
};

export type StartSessionInput = {
  surveyId: string;
  shareLinkId?: string;
  clientMetadata: Record<string, unknown>;
};

export type StartSessionResult = {
  sessionId: string;
  firstQuestion: QuestionView;
};

export type SubmitAnswerInput = {
  sessionId: string;
  questionId: string;
  answerId?: string;
  answerIds?: string[];
  rawText?: string;
};

export type ReportDelivery = {
  status: ReportDeliveryStatus;
  email: string;
  fileName: string;
  contentType: string;
  errorMessage?: string;
};

export type SubmitAnswerResult = {
  nextQuestionId: string;
  isFinished: boolean;
  nextQuestion?: QuestionView;
  reportDelivery?: ReportDelivery;
};

export type SessionResponse = {
  questionId: string;
  questionText: string;
  questionType: QuestionType;
  answerId?: string;
  answerIds?: string[];
  rawText?: string;
  selectedWeight?: number;
  categoryTag?: string;
};

export type SessionAnalytics = {
  surveyId: string;
  sessionId: string;
  clientMetadata: Record<string, unknown>;
  responses: SessionResponse[];
  startedAt: string;
  finishedAt?: string;
};

export type SessionRecord = {
  sessionId: string;
  surveyId: string;
  shareLinkId?: string;
  clientMetadata: Record<string, unknown>;
  currentQuestionId: string;
  currentIndex: number;
  status: SessionStatus;
  startedAt: string;
  finishedAt?: string;
  responsesCount?: number;
  responses: SessionResponse[];
  reportDelivery?: ReportDelivery;
};

export type ReportPreview = {
  title: string;
  subtitle: string;
  summary: string[];
  strengths: string[];
  developmentPoints: string[];
  responseHighlights: string[];
};

export type AppSession = {
  tokens: AuthTokens;
  profile: UserProfile;
};

export type CandidateSetup = {
  shareLinkId?: string;
  title: string;
  description: string;
  intro: string;
  ownerId?: string;
  fields: StartFormField[];
};
