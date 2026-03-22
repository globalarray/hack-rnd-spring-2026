import { FormEvent, useEffect, useMemo, useState } from "react";
import { Navigate } from "react-router-dom";

import { useAuth } from "../app/auth";
import { SurveyConstructor } from "../components/survey-constructor";
import { WorkspaceLayout, WorkspaceNav, LogoutAction } from "../components/workspace";
import {
  Badge,
  Button,
  Card,
  EmptyState,
  Field,
  GhostButton,
  Input,
  Modal,
  SectionTitle,
  StatCard,
  Tabs,
  TextArea
} from "../components/ui";
import { api } from "../lib/api";
import { buildReportPreview } from "../lib/reports";
import type {
  ReportPreview,
  SessionRecord,
  ShareLinkConfig,
  SurveyDraft,
  SurveyRecord,
  SurveySummary
} from "../lib/types";
import { buildDefaultSettings, buildQuestionTemplate, formatDate, formatDateTime, readErrorMessage } from "../lib/utils";

type PsychologistTab = "tests" | "constructor" | "links" | "results" | "profile";

function buildStarterDraft(psychologistId: string): SurveyDraft {
  return {
    psychologistId,
    title: "Новая методика",
    description: "Черновик теста для психолога. Здесь можно собрать flow, metadata и сценарии переходов.",
    settings: buildDefaultSettings(),
    questions: [
      buildQuestionTemplate("single_choice", 1),
      buildQuestionTemplate("text", 2)
    ]
  };
}

function surveyToDraft(survey: SurveyRecord): SurveyDraft {
  return {
    surveyId: survey.surveyId,
    psychologistId: survey.psychologistId,
    title: survey.title,
    description: survey.description,
    settings: survey.settings,
    questions: survey.questions
  };
}

export function PsychologistDashboard() {
  const { session, logout, updateProfile } = useAuth();
  const [activeTab, setActiveTab] = useState<PsychologistTab>("tests");
  const [surveys, setSurveys] = useState<SurveySummary[]>([]);
  const [selectedSurveyId, setSelectedSurveyId] = useState("");
  const [activeSurvey, setActiveSurvey] = useState<SurveyRecord | null>(null);
  const [currentDraft, setCurrentDraft] = useState<SurveyDraft | null>(null);
  const [sessions, setSessions] = useState<SessionRecord[]>([]);
  const [profileDraft, setProfileDraft] = useState({
    about: session?.profile.about ?? "",
    photoUrl: session?.profile.photoUrl ?? ""
  });
  const [linkDraft, setLinkDraft] = useState({
    title: "",
    description: "",
    intro: ""
  });
  const [selectedFieldKeys, setSelectedFieldKeys] = useState<string[]>([]);
  const [feedback, setFeedback] = useState("");
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSavingSurvey, setIsSavingSurvey] = useState(false);
  const [isSavingProfile, setIsSavingProfile] = useState(false);
  const [reportPreview, setReportPreview] = useState<ReportPreview | null>(null);

  useEffect(() => {
    if (!session) {
      return;
    }

    setCurrentDraft((current) => current ?? buildStarterDraft(session.profile.id));
    setProfileDraft({
      about: session.profile.about,
      photoUrl: session.profile.photoUrl
    });
  }, [session]);

  useEffect(() => {
    if (!session) {
      return;
    }

    setIsLoading(true);
    api
      .listSurveys(session.tokens.accessToken, session.profile.id)
      .then((items) => {
        setSurveys(items);
        setSelectedSurveyId((current) => current || items[0]?.surveyId || "");
      })
      .catch((loadError) => setError(readErrorMessage(loadError)))
      .finally(() => setIsLoading(false));
  }, [session]);

  useEffect(() => {
    if (!session || !selectedSurveyId) {
      setActiveSurvey(null);
      setSessions([]);
      return;
    }

    api
      .getSurvey(session.tokens.accessToken, selectedSurveyId)
      .then((survey) => {
        setActiveSurvey(survey);
        setSelectedFieldKeys(survey.settings.startForm.fields.slice(2).map((field) => field.key));
        setLinkDraft({
          title: `Ссылка для ${survey.title}`,
          description: "Стандартный запуск теста",
          intro: survey.settings.startForm.intro
        });
      })
      .catch(() => {
        setActiveSurvey(null);
      });

    api
      .listSurveySessions(session.tokens.accessToken, selectedSurveyId)
      .then(setSessions)
      .catch(() => setSessions([]));
  }, [selectedSurveyId, session]);

  const stats = useMemo(() => {
    const activeTests = surveys.filter((survey) => survey.status === "active").length;
    const totalCompletions = surveys.reduce((sum, survey) => sum + survey.completionsCount, 0);
    return {
      totalTests: surveys.length,
      activeTests,
      totalCompletions,
      totalLinks: activeSurvey?.shareLinks.length ?? 0
    };
  }, [activeSurvey?.shareLinks.length, surveys]);

  if (!session) {
    return null;
  }

  if (session.profile.role !== "psychologist") {
    return <Navigate to="/admin" replace />;
  }

  async function refreshSurveys(nextSelectedSurveyId?: string) {
    const items = await api.listSurveys(session.tokens.accessToken, session.profile.id);
    setSurveys(items);
    const resolvedId = nextSelectedSurveyId ?? selectedSurveyId ?? items[0]?.surveyId ?? "";
    setSelectedSurveyId(resolvedId);
    if (resolvedId) {
      try {
        const survey = await api.getSurvey(session.tokens.accessToken, resolvedId);
        setActiveSurvey(survey);
      } catch {
        setActiveSurvey(null);
      }
    }
  }

  async function handleSaveSurvey() {
    if (!currentDraft) {
      return;
    }

    setFeedback("");
    setError("");
    setIsSavingSurvey(true);

    try {
      if (currentDraft.surveyId) {
        await api.updateSurvey(session.tokens.accessToken, currentDraft.surveyId, currentDraft);
        setFeedback("Тест обновлен и остался в конструкторе.");
        await refreshSurveys(currentDraft.surveyId);
      } else {
        const response = await api.createSurvey(session.tokens.accessToken, currentDraft);
        const survey = await api.getSurvey(session.tokens.accessToken, response.surveyId).catch(() => null);
        setFeedback("Тест сохранен. Теперь можно создавать публичные ссылки и смотреть прохождения.");
        await refreshSurveys(response.surveyId);
        setCurrentDraft(survey ? surveyToDraft(survey) : { ...currentDraft, surveyId: response.surveyId });
      }
    } catch (saveError) {
      setError(readErrorMessage(saveError));
    } finally {
      setIsSavingSurvey(false);
    }
  }

  function openSurveyInConstructor(surveyId: string) {
    setSelectedSurveyId(surveyId);
    setActiveTab("constructor");
    if (!session) {
      return;
    }

    api
      .getSurvey(session.tokens.accessToken, surveyId)
      .then((survey) => setCurrentDraft(surveyToDraft(survey)))
      .catch((loadError) => setError(readErrorMessage(loadError)));
  }

  function startNewDraft() {
    setCurrentDraft(buildStarterDraft(session.profile.id));
    setSelectedSurveyId("");
    setActiveSurvey(null);
    setActiveTab("constructor");
  }

  async function handleAnnulSurvey(surveyId: string) {
    setFeedback("");
    setError("");

    try {
      await api.annulSurvey(session.tokens.accessToken, surveyId);
      await refreshSurveys(surveyId);
      setFeedback("Тест аннулирован. Новые прохождения по нему закрыты.");
    } catch (annulError) {
      setError(readErrorMessage(annulError));
    }
  }

  async function handleCreateLink(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!activeSurvey) {
      return;
    }

    setFeedback("");
    setError("");

    try {
      const extraFields = activeSurvey.settings.startForm.fields.filter(
        (field, index) => index >= 2 && selectedFieldKeys.includes(field.key)
      );
      const link = await api.createShareLink(session.tokens.accessToken, activeSurvey.surveyId, {
        ...linkDraft,
        extraFields,
        allowSelfReport: true
      });

      setActiveSurvey({
        ...activeSurvey,
        shareLinks: [link, ...activeSurvey.shareLinks]
      });
      setFeedback("Новая публичная ссылка готова. Ее можно сразу отправлять участнику.");
    } catch (linkError) {
      setError(readErrorMessage(linkError));
    }
  }

  async function handleCopy(value: string) {
    await navigator.clipboard.writeText(value);
    setFeedback("Ссылка скопирована.");
  }

  async function handleSendClientReport(sessionId: string) {
    setFeedback("");
    setError("");

    try {
      const result = await api.sendSessionReport(session.tokens.accessToken, sessionId, "client_docx");
      setFeedback(`Отчет отправлен на ${result.email}.`);
    } catch (reportError) {
      setError(readErrorMessage(reportError));
    }
  }

  async function handleOpenPsychologistReport(sessionId: string) {
    if (!activeSurvey) {
      return;
    }

    setFeedback("");
    setError("");

    try {
      const analytics = await api.getSessionAnalytics(session.tokens.accessToken, sessionId);
      setReportPreview(buildReportPreview(activeSurvey, analytics));
      await api.sendSessionReport(session.tokens.accessToken, sessionId, "psycho_docx");
      setFeedback("Личный отчет сформирован: открыт в окне и дополнительно отправлен вам на email.");
    } catch (reportError) {
      setError(readErrorMessage(reportError));
    }
  }

  async function handleProfileSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback("");
    setError("");
    setIsSavingProfile(true);

    try {
      await updateProfile(profileDraft);
      setFeedback("Профиль психолога обновлен.");
    } catch (submitError) {
      setError(readErrorMessage(submitError));
    } finally {
      setIsSavingProfile(false);
    }
  }

  const selectedFields = activeSurvey?.settings.startForm.fields.slice(2) ?? [];

  return (
    <WorkspaceLayout
      title="Панель психолога"
      subtitle="Здесь собраны ваши методики, drag-and-drop конструктор, публичные ссылки на прохождение и результаты с email-отчетами."
      badge="Psychologist"
      aside={(
        <WorkspaceNav
          title="Psychologist Console"
          subtitle="Тесты, сессии, ссылки и аналитика"
          value={activeTab}
          options={[
            { value: "tests", label: "Тесты", description: "Мои методики и статусы" },
            { value: "constructor", label: "Конструктор", description: "Редактирование вопросов и логики" },
            { value: "links", label: "Ссылки", description: "Публичные сценарии запуска" },
            { value: "results", label: "Результаты", description: "Прошедшие и отчеты" },
            { value: "profile", label: "Профиль", description: "Публичная карточка психолога" }
          ]}
          onChange={setActiveTab}
          footer={<LogoutAction onLogout={logout} />}
        />
      )}
      topActions={(
        <>
          <Badge>{api.mode === "mock" ? "Demo data" : "Live BFF"}</Badge>
          <Button type="button" onClick={startNewDraft}>
            Новый тест
          </Button>
        </>
      )}
    >
      <section className="stats-grid">
        <StatCard label="Всего тестов" value={stats.totalTests} hint="Активные и аннулированные методики" />
        <StatCard label="Активные тесты" value={stats.activeTests} hint="Доступны для новых прохождений" />
        <StatCard label="Завершенные сессии" value={stats.totalCompletions} hint="Сумма прохождений по всем тестам" />
        <StatCard label="Ссылки текущего теста" value={stats.totalLinks} hint="Можно создавать неограниченно" />
      </section>

      {feedback ? <p className="feedback feedback--success">{feedback}</p> : null}
      {error ? <p className="feedback feedback--error">{error}</p> : null}

      {activeTab === "tests" ? (
        <div className="stack">
          <SectionTitle
            title="Мои тесты"
            description="Для каждого теста видно число прохождений, статус, быстрый переход в конструктор и в результаты."
            action={<Button type="button" onClick={startNewDraft}>Создать новый тест</Button>}
          />

          {isLoading ? (
            <p className="muted">Загружаем тесты...</p>
          ) : surveys.length === 0 ? (
            <EmptyState
              title="Тестов пока нет"
              description="Соберите первую методику в конструкторе и сохраните ее в BFF."
              action={<Button type="button" onClick={startNewDraft}>Открыть конструктор</Button>}
            />
          ) : (
            <div className="table-list">
              {surveys.map((survey) => (
                <div key={survey.surveyId} className="table-row">
                  <div>
                    <strong>{survey.title}</strong>
                    <span>ID {survey.surveyId}</span>
                  </div>
                  <div>
                    <span>Прохождений: {survey.completionsCount}</span>
                    <Badge className={`status-badge status-badge--${survey.status}`}>{survey.status}</Badge>
                  </div>
                  <div className="table-row__actions">
                    <GhostButton type="button" onClick={() => { setSelectedSurveyId(survey.surveyId); setActiveTab("results"); }}>
                      Открыть результаты
                    </GhostButton>
                    <GhostButton type="button" onClick={() => { setSelectedSurveyId(survey.surveyId); setActiveTab("links"); }}>
                      Ссылки
                    </GhostButton>
                    <Button type="button" onClick={() => openSurveyInConstructor(survey.surveyId)}>
                      Редактировать
                    </Button>
                    <GhostButton type="button" onClick={() => handleAnnulSurvey(survey.surveyId)}>
                      Аннулировать
                    </GhostButton>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      ) : null}

      {activeTab === "constructor" ? (
        currentDraft ? (
          <SurveyConstructor
            draft={currentDraft}
            onChange={setCurrentDraft}
            onSave={handleSaveSurvey}
            isSaving={isSavingSurvey}
            shareLinks={activeSurvey?.shareLinks ?? []}
          />
        ) : (
          <EmptyState
            title="Черновик не выбран"
            description="Откройте существующий тест или начните новый черновик."
            action={<Button type="button" onClick={startNewDraft}>Создать черновик</Button>}
          />
        )
      ) : null}

      {activeTab === "links" ? (
        <div className="workspace-grid">
          <Card className="stack">
            <SectionTitle
              title="Sessions / public links"
              description="Каждая ссылка может запрашивать свой набор metadata перед стартом теста. ФИО и email всегда остаются обязательными."
            />

            {!activeSurvey ? (
              <EmptyState
                title="Выберите тест"
                description="Сначала откройте методику из списка тестов, после чего здесь появится настройка ссылок."
              />
            ) : (
              <form className="form-grid" onSubmit={handleCreateLink}>
                <Field label="Название ссылки">
                  <Input
                    value={linkDraft.title}
                    onChange={(event) => setLinkDraft((state) => ({ ...state, title: event.target.value }))}
                    placeholder="Например: 11 класс, поток А"
                  />
                </Field>
                <Field label="Описание">
                  <TextArea
                    rows={3}
                    value={linkDraft.description}
                    onChange={(event) => setLinkDraft((state) => ({ ...state, description: event.target.value }))}
                    placeholder="Короткий комментарий для внутренней навигации"
                  />
                </Field>
                <Field label="Вступительный текст">
                  <TextArea
                    rows={4}
                    value={linkDraft.intro}
                    onChange={(event) => setLinkDraft((state) => ({ ...state, intro: event.target.value }))}
                  />
                </Field>
                <Card className="stack highlight-card">
                  <strong>Дополнительная metadata</strong>
                  {selectedFields.length === 0 ? (
                    <span>Добавьте поля в конструкторе, чтобы переиспользовать их в ссылках.</span>
                  ) : (
                    <div className="checkbox-grid">
                      {selectedFields.map((field) => (
                        <label key={field.key} className="checkbox-inline checkbox-card">
                          <input
                            type="checkbox"
                            checked={selectedFieldKeys.includes(field.key)}
                            onChange={(event) => {
                              if (event.target.checked) {
                                setSelectedFieldKeys((current) => [...current, field.key]);
                              } else {
                                setSelectedFieldKeys((current) => current.filter((item) => item !== field.key));
                              }
                            }}
                          />
                          <span>{field.label}</span>
                        </label>
                      ))}
                    </div>
                  )}
                </Card>
                <Button type="submit">Создать ссылку</Button>
              </form>
            )}
          </Card>

          <Card className="stack">
            <SectionTitle
              title="Готовые ссылки"
              description="Каждая ссылка открывает public flow на `site_url/tests/{surveyId}/start` с собственным набором metadata."
            />

            {!activeSurvey || activeSurvey.shareLinks.length === 0 ? (
              <EmptyState
                title="Пока нет ссылок"
                description="Сохраните тест и создайте первый публичный entry point."
              />
            ) : (
              <div className="table-list">
                {activeSurvey.shareLinks.map((link: ShareLinkConfig) => (
                  <div key={link.id} className="table-row">
                    <div>
                      <strong>{link.title}</strong>
                      <span>{link.description}</span>
                    </div>
                    <div>
                      <span>Создано {formatDate(link.createdAt)}</span>
                      <span>{link.extraFields.length} доп. полей</span>
                    </div>
                    <div className="table-row__actions">
                      <GhostButton type="button" onClick={() => window.open(link.publicUrl, "_blank", "noopener,noreferrer")}>
                        Открыть
                      </GhostButton>
                      <Button type="button" onClick={() => handleCopy(link.publicUrl)}>
                        Скопировать ссылку
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </div>
      ) : null}

      {activeTab === "results" ? (
        <div className="stack">
          <SectionTitle
            title="Прошедшие и отчеты"
            description="По каждому прохождению можно отправить клиентский отчет на email и открыть психологический отчет для себя."
          />

          {surveys.length > 0 ? (
            <Tabs
              value={selectedSurveyId as string}
              options={surveys.map((survey) => ({ value: survey.surveyId, label: survey.title }))}
              onChange={setSelectedSurveyId}
            />
          ) : null}

          {!selectedSurveyId ? (
            <EmptyState
              title="Нет выбранного теста"
              description="Выберите методику в списке тестов, чтобы смотреть прохождения."
            />
          ) : sessions.length === 0 ? (
            <EmptyState
              title="Прохождений пока нет"
              description="Как только по ссылке пройдет первый участник, здесь появится его карточка, кнопки email-отчетов и личный разбор."
            />
          ) : (
            <div className="table-list">
              {sessions.map((sessionRow) => (
                <div key={sessionRow.sessionId} className="table-row table-row--dense">
                  <div>
                    <strong>{String(sessionRow.clientMetadata.fullName ?? "Без имени")}</strong>
                    <span>{String(sessionRow.clientMetadata.email ?? "Email не указан")}</span>
                  </div>
                  <div>
                    <span>Завершено {formatDateTime(sessionRow.finishedAt)}</span>
                    <span>{sessionRow.responses.length} ответов</span>
                  </div>
                  <div className="table-row__actions">
                    <GhostButton type="button" onClick={() => handleOpenPsychologistReport(sessionRow.sessionId)}>
                      Сформировать для себя
                    </GhostButton>
                    <Button type="button" onClick={() => handleSendClientReport(sessionRow.sessionId)}>
                      Отправить клиенту
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      ) : null}

      {activeTab === "profile" ? (
        <Card className="stack">
          <SectionTitle
            title="Публичный профиль психолога"
            description="Эти данные можно показывать в карточке специалиста рядом с публичной ссылкой на тест."
          />

          <form className="form-grid" onSubmit={handleProfileSubmit}>
            <Field label="ФИО">
              <Input value={session.profile.fullName} disabled />
            </Field>
            <Field label="Email">
              <Input value={session.profile.email} disabled />
            </Field>
            <Field label="Ссылка на фото">
              <Input
                value={profileDraft.photoUrl}
                onChange={(event) => setProfileDraft((state) => ({ ...state, photoUrl: event.target.value }))}
                placeholder="https://example.com/avatar.jpg"
              />
            </Field>
            <Field label="О себе">
              <TextArea
                rows={6}
                value={profileDraft.about}
                onChange={(event) => setProfileDraft((state) => ({ ...state, about: event.target.value }))}
                placeholder="Опишите специализацию, подход и для какой аудитории вы проводите диагностику."
              />
            </Field>

            <Button type="submit" disabled={isSavingProfile}>
              {isSavingProfile ? "Сохраняем..." : "Сохранить профиль"}
            </Button>
          </form>
        </Card>
      ) : null}

      <Modal isOpen={Boolean(reportPreview)} title={reportPreview?.title ?? ""} onClose={() => setReportPreview(null)}>
        {reportPreview ? (
          <div className="report-preview">
            <p className="report-preview__subtitle">{reportPreview.subtitle}</p>
            <section>
              <h4>Резюме</h4>
              <ul>
                {reportPreview.summary.map((item) => <li key={item}>{item}</li>)}
              </ul>
            </section>
            <section>
              <h4>Сильные стороны</h4>
              <ul>
                {reportPreview.strengths.map((item) => <li key={item}>{item}</li>)}
              </ul>
            </section>
            <section>
              <h4>Зоны уточнения</h4>
              <ul>
                {reportPreview.developmentPoints.map((item) => <li key={item}>{item}</li>)}
              </ul>
            </section>
            <section>
              <h4>Ключевые ответы</h4>
              <ul>
                {reportPreview.responseHighlights.map((item) => <li key={item}>{item}</li>)}
              </ul>
            </section>
          </div>
        ) : null}
      </Modal>
    </WorkspaceLayout>
  );
}
