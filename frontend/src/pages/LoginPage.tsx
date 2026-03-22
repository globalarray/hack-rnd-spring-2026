import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";

import { BrandHomeLink } from "../components/brand";
import { Badge, Button, Card, Field, GhostButton, Input } from "../components/ui";
import { api } from "../lib/api";
import { readErrorMessage } from "../lib/utils";
import { useAuth } from "../app/auth";

export function LoginPage() {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [email, setEmail] = useState("admin@profdnk.local");
  const [password, setPassword] = useState("admin12345");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const session = await login(email, password);
      navigate(session.profile.role === "admin" ? "/admin" : "/psychologist", { replace: true });
    } catch (submitError) {
      setError(readErrorMessage(submitError));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="auth-page">
      <BrandHomeLink compact className="brand-floating" />
      <section className="auth-page__hero">
        <Badge>{api.mode === "mock" ? "Demo mode" : "BFF mode"}</Badge>
        <h1>Одна точка входа для администратора и психолога.</h1>
        <p>
          После авторизации мы сами откроем нужный кабинет: администратор увидит управление приглашениями,
          а психолог сразу попадет в тесты, ссылки и результаты.
        </p>
        <div className="auth-page__feature-list">
          <Card>
            <strong>Admin workspace</strong>
            <span>Создание профилей психологов, блокировка, выдача invitation-ссылок.</span>
          </Card>
          <Card>
            <strong>Psychologist console</strong>
            <span>Конструктор методик, публичные ссылки, сессии, email-отчеты и личные отчеты.</span>
          </Card>
          <Card>
            <strong>Candidate flow</strong>
            <span>Простая анкета перед стартом, поэтапное прохождение и финальная отправка отчета на почту.</span>
          </Card>
        </div>
      </section>

      <Card className="auth-card">
        <div className="auth-card__header">
          <div>
            <p className="eyebrow">Sign in</p>
            <h2>Вход в систему</h2>
          </div>
          <Badge>Unified login</Badge>
        </div>

        <form className="auth-form" onSubmit={handleSubmit}>
          <Field label="Email">
            <Input
              autoComplete="email"
              value={email}
              onChange={(event) => setEmail(event.target.value)}
              placeholder="admin@profdnk.local"
            />
          </Field>
          <Field label="Пароль">
            <Input
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="Введите пароль"
            />
          </Field>

          {error ? <p className="feedback feedback--error">{error}</p> : null}

          <Button type="submit" disabled={isSubmitting}>
            {isSubmitting ? "Проверяем доступ..." : "Войти"}
          </Button>
        </form>

        <div className="login-hints">
          <span>Для mock-режима уже готовы тестовые входы:</span>
          <GhostButton type="button" onClick={() => { setEmail("admin@profdnk.local"); setPassword("admin12345"); }}>
            Администратор
          </GhostButton>
          <GhostButton type="button" onClick={() => { setEmail("psycho@profdnk.local"); setPassword("psych12345"); }}>
            Психолог
          </GhostButton>
        </div>
      </Card>
    </main>
  );
}
