import { FormEvent, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { useAuth } from "../app/auth";
import { BrandHomeLink } from "../components/brand";
import { Badge, Button, Card, Field, Input } from "../components/ui";
import { readErrorMessage } from "../lib/utils";

export function InvitationPage() {
  const { token = "" } = useParams();
  const navigate = useNavigate();
  const { register } = useAuth();
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      await register(token, password);
      navigate("/psychologist", { replace: true });
    } catch (submitError) {
      setError(readErrorMessage(submitError));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <main className="invitation-page">
      <BrandHomeLink compact className="brand-floating" />
      <Card className="invitation-card">
        <Badge>Invitation</Badge>
        <h1>Завершение регистрации психолога</h1>
        <p>
          Администратор уже заполнил ваши данные. На этом экране нужен только пароль, после чего кабинет
          откроется автоматически.
        </p>

        <form className="auth-form" onSubmit={handleSubmit}>
          <Field
            label="Создайте пароль"
            hint="Минимум 8 символов. Остальные данные уже закреплены за приглашением."
          >
            <Input
              type="password"
              autoComplete="new-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="Надежный пароль"
            />
          </Field>

          {error ? <p className="feedback feedback--error">{error}</p> : null}

          <Button type="submit" disabled={isSubmitting || password.trim().length < 8}>
            {isSubmitting ? "Создаем кабинет..." : "Подтвердить и войти"}
          </Button>
        </form>
      </Card>
    </main>
  );
}
