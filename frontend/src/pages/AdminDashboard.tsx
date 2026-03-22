import { FormEvent, useEffect, useMemo, useState } from "react";
import { Navigate } from "react-router-dom";

import { useAuth } from "../app/auth";
import { WorkspaceLayout, WorkspaceNav, LogoutAction } from "../components/workspace";
import {
  Badge,
  Button,
  Card,
  EmptyState,
  Field,
  GhostButton,
  Input,
  SectionTitle,
  StatCard,
  TextArea
} from "../components/ui";
import { api } from "../lib/api";
import type { DirectoryItem } from "../lib/types";
import { copyText, formatDate, readErrorMessage } from "../lib/utils";

type AdminTab = "directory" | "profile";

const defaultInviteForm = {
  fullName: "",
  phone: "",
  email: "",
  accessUntil: "",
  expiresAt: ""
};

function todayDateInputValue() {
  const now = new Date();
  const offsetMinutes = now.getTimezoneOffset();
  const localDate = new Date(now.getTime() - offsetMinutes * 60_000);
  return localDate.toISOString().slice(0, 10);
}

export function AdminDashboard() {
  const { session, logout, updateProfile } = useAuth();
  const [activeTab, setActiveTab] = useState<AdminTab>("directory");
  const [directory, setDirectory] = useState<DirectoryItem[]>([]);
  const [inviteForm, setInviteForm] = useState(defaultInviteForm);
  const [profileDraft, setProfileDraft] = useState({
    about: session?.profile.about ?? "",
    photoUrl: session?.profile.photoUrl ?? ""
  });
  const [createdLink, setCreatedLink] = useState("");
  const [feedback, setFeedback] = useState("");
  const [error, setError] = useState("");
  const [copiedLink, setCopiedLink] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const minDate = useMemo(() => todayDateInputValue(), []);

  useEffect(() => {
    setProfileDraft({
      about: session?.profile.about ?? "",
      photoUrl: session?.profile.photoUrl ?? ""
    });
  }, [session?.profile.about, session?.profile.photoUrl]);

  useEffect(() => {
    if (!session) {
      return;
    }

    setIsLoading(true);
    api
      .listPsychologists(session.tokens.accessToken)
      .then(setDirectory)
      .catch((loadError) => setError(readErrorMessage(loadError)))
      .finally(() => setIsLoading(false));
  }, [session]);

  useEffect(() => {
    if (!copiedLink) {
      return;
    }

    const timeoutId = window.setTimeout(() => setCopiedLink(""), 3000);
    return () => window.clearTimeout(timeoutId);
  }, [copiedLink]);

  const summary = useMemo(() => {
    const registered = directory.filter((item) => item.status !== "pending");
    return {
      total: directory.length,
      active: registered.filter((item) => item.status === "active").length,
      blocked: registered.filter((item) => item.status === "blocked").length,
      pending: directory.filter((item) => item.status === "pending").length
    };
  }, [directory]);

  if (!session) {
    return null;
  }

  if (session.profile.role !== "admin") {
    return <Navigate to="/psychologist" replace />;
  }

  async function refreshDirectory() {
    const nextDirectory = await api.listPsychologists(session.tokens.accessToken);
    setDirectory(nextDirectory);
  }

  async function handleInviteSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback("");
    setError("");
    setIsSubmitting(true);

    try {
      const response = await api.createInvitation(session.tokens.accessToken, {
        ...inviteForm,
        role: "psychologist"
      });
      setCreatedLink(response.invitationUrl);
      setFeedback("Приглашение создано. Ссылку можно сразу отправлять психологу.");
      setInviteForm(defaultInviteForm);
      await refreshDirectory();
    } catch (submitError) {
      setError(readErrorMessage(submitError));
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleToggleBlock(item: DirectoryItem) {
    setFeedback("");
    setError("");

    try {
      if (item.status === "blocked") {
        await api.unblockUser(session.tokens.accessToken, item.email);
        setFeedback(`Психолог ${item.email} разблокирован.`);
      } else {
        await api.blockUser(session.tokens.accessToken, item.email);
        setFeedback(`Психолог ${item.email} заблокирован.`);
      }
      await refreshDirectory();
    } catch (actionError) {
      setError(readErrorMessage(actionError));
    }
  }

  async function handleCopyLink(value: string) {
    setFeedback("");
    setError("");

    try {
      await copyText(value);
      setCopiedLink(value);
      setFeedback("Ссылка скопирована в буфер обмена.");
    } catch (copyError) {
      setError(readErrorMessage(copyError));
    }
  }

  async function handleProfileSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setFeedback("");
    setError("");
    setIsSubmitting(true);

    try {
      await updateProfile(profileDraft);
      setFeedback("Профиль администратора обновлен.");
    } catch (submitError) {
      setError(readErrorMessage(submitError));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <WorkspaceLayout
      title="Административная панель"
      subtitle="Единая зона для выдачи доступов, контроля статусов и запуска новых психологов через invitation-flow."
      badge="Admin"
      aside={(
        <WorkspaceNav
          title="Admin Console"
          subtitle="Приглашения, статусы и bootstrap-доступы"
          value={activeTab}
          options={[
            {
              value: "directory",
              label: "Психологи",
              description: "Создание профилей, ссылки и блокировка"
            },
            {
              value: "profile",
              label: "Профиль",
              description: "Информация администратора и описание кабинета"
            }
          ]}
          onChange={setActiveTab}
          footer={<LogoutAction onLogout={logout} />}
        />
      )}
      topActions={<Badge>{api.mode === "mock" ? "Demo data" : "Live BFF"}</Badge>}
    >
      <section className="stats-grid">
        <StatCard label="Всего карточек" value={summary.total} hint="Зарегистрированные и ожидающие приглашения" />
        <StatCard label="Активные" value={summary.active} hint="Могут входить и работать с тестами" />
        <StatCard label="Заблокированные" value={summary.blocked} hint="Доступ временно закрыт" />
        <StatCard label="Ожидают пароль" value={summary.pending} hint="Получили ссылку, но еще не завершили регистрацию" />
      </section>

      {feedback ? <p className="feedback feedback--success">{feedback}</p> : null}
      {error ? <p className="feedback feedback--error">{error}</p> : null}

      {activeTab === "directory" ? (
        <div className="workspace-grid">
          <Card className="stack">
            <SectionTitle
              title="Создать профиль психолога"
              description="Администратор заполняет данные заранее, а психолог потом вводит только пароль по ссылке."
            />

            <form className="form-grid" onSubmit={handleInviteSubmit}>
              <Field label="ФИО">
                <Input
                  value={inviteForm.fullName}
                  onChange={(event) => setInviteForm((state) => ({ ...state, fullName: event.target.value }))}
                  placeholder="Анна Смирнова"
                />
              </Field>
              <Field label="Телефон">
                <Input
                  value={inviteForm.phone}
                  onChange={(event) => setInviteForm((state) => ({ ...state, phone: event.target.value }))}
                  placeholder="+79990001122"
                />
              </Field>
              <Field label="Email">
                <Input
                  type="email"
                  value={inviteForm.email}
                  onChange={(event) => setInviteForm((state) => ({ ...state, email: event.target.value }))}
                  placeholder="anna@example.com"
                />
              </Field>
              <Field label="Доступ до">
                <Input
                  type="date"
                  min={minDate}
                  value={inviteForm.accessUntil}
                  onChange={(event) =>
                    setInviteForm((state) => {
                      const accessUntil = event.target.value;
                      const nextExpiresAt = !state.expiresAt || state.expiresAt > accessUntil
                        ? accessUntil
                        : state.expiresAt;

                      return {
                        ...state,
                        accessUntil,
                        expiresAt: nextExpiresAt
                      };
                    })
                  }
                />
              </Field>
              <Field
                label="Ссылка действительна до"
                hint="Ссылка будет активна до конца выбранного дня. Время указывать не нужно."
              >
                <Input
                  type="date"
                  min={minDate}
                  max={inviteForm.accessUntil || undefined}
                  value={inviteForm.expiresAt}
                  onChange={(event) => setInviteForm((state) => ({ ...state, expiresAt: event.target.value }))}
                />
              </Field>

              <Button type="submit" disabled={isSubmitting}>
                {isSubmitting ? "Создаем..." : "Создать приглашение"}
              </Button>
            </form>

            {createdLink ? (
              <Card className="highlight-card">
                <strong>Invitation link</strong>
                <p>{createdLink}</p>
                <GhostButton
                  type="button"
                  className={copiedLink === createdLink ? "is-copied" : undefined}
                  disabled={copiedLink === createdLink}
                  onClick={() => handleCopyLink(createdLink)}
                >
                  {copiedLink === createdLink ? "Скопировано" : "Скопировать ссылку"}
                </GhostButton>
              </Card>
            ) : null}
          </Card>

          <Card className="stack">
            <SectionTitle
              title="Каталог психологов"
              description="Здесь видно, кто уже зарегистрирован, кто ждет пароль и кого можно временно блокировать."
            />

            {isLoading ? (
              <p className="muted">Загружаем каталог...</p>
            ) : directory.length === 0 ? (
              <EmptyState
                title="Пока нет специалистов"
                description="Как только вы создадите первое приглашение, здесь появится карточка психолога."
              />
            ) : (
              <div className="table-list">
                {directory.map((item) => (
                  <div key={item.email} className="table-row">
                    <div>
                      <strong>{item.fullName}</strong>
                      <span>{item.email}</span>
                    </div>
                    <div>
                      <span>{item.phone || "Телефон пока не указан"}</span>
                      <span>Доступ до {formatDate(item.accessUntil)}</span>
                    </div>
                    <div className="table-row__actions">
                      <Badge className={`status-badge status-badge--${item.status}`}>{item.status}</Badge>
                      {item.invitationUrl ? (
                        <GhostButton
                          type="button"
                          className={copiedLink === item.invitationUrl ? "is-copied" : undefined}
                          disabled={copiedLink === item.invitationUrl}
                          onClick={() => handleCopyLink(item.invitationUrl ?? "")}
                        >
                          {copiedLink === item.invitationUrl ? "Скопировано" : "Скопировать инвайт"}
                        </GhostButton>
                      ) : null}
                      {item.status !== "pending" ? (
                        <Button type="button" onClick={() => handleToggleBlock(item)}>
                          {item.status === "blocked" ? "Разблокировать" : "Заблокировать"}
                        </Button>
                      ) : null}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </Card>
        </div>
      ) : (
        <Card className="stack">
          <SectionTitle
            title="Профиль администратора"
            description="Короткое описание роли и контактная информация для системных сценариев."
          />

          <form className="form-grid" onSubmit={handleProfileSubmit}>
            <Field label="Полное имя">
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
                rows={5}
                value={profileDraft.about}
                onChange={(event) => setProfileDraft((state) => ({ ...state, about: event.target.value }))}
                placeholder="Коротко опишите роль администратора в платформе."
              />
            </Field>

            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Сохраняем..." : "Сохранить профиль"}
            </Button>
          </form>
        </Card>
      )}
    </WorkspaceLayout>
  );
}
