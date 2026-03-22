import { FormEvent, useEffect, useMemo, useState } from "react";
import { useLocation, useParams } from "react-router-dom";

import { BrandHomeLink } from "../components/brand";
import { Badge, Button, Card, Field, Input, TextArea } from "../components/ui";
import { api } from "../lib/api";
import type { CandidateSetup, QuestionView, ReportDelivery } from "../lib/types";
import { DEFAULT_FIELDS, decodeSetup, readErrorMessage } from "../lib/utils";

function mergeFields(setup: CandidateSetup | null) {
  const extras = setup?.fields ?? [];
  const keys = new Set(DEFAULT_FIELDS.map((field) => field.key));
  return [
    ...DEFAULT_FIELDS,
    ...extras.filter((field) => !keys.has(field.key))
  ];
}

export function CandidateFlowPage() {
  const { surveyId = "" } = useParams();
  const location = useLocation();
  const setup = useMemo(() => decodeSetup(new URLSearchParams(location.search).get("setup")), [location.search]);
  const fields = useMemo(() => mergeFields(setup), [setup]);
  const [metadata, setMetadata] = useState<Record<string, string>>({});
  const [sessionId, setSessionId] = useState("");
  const [currentQuestion, setCurrentQuestion] = useState<QuestionView | null>(null);
  const [currentAnswerId, setCurrentAnswerId] = useState("");
  const [currentAnswerIds, setCurrentAnswerIds] = useState<string[]>([]);
  const [rawText, setRawText] = useState("");
  const [reportDelivery, setReportDelivery] = useState<ReportDelivery | null>(null);
  const [phase, setPhase] = useState<"start" | "testing" | "finished">("start");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    const initialState = Object.fromEntries(fields.map((field) => [field.key, ""]));
    setMetadata(initialState);
  }, [fields]);

  useEffect(() => {
    setCurrentAnswerId("");
    setCurrentAnswerIds([]);
    setRawText("");
  }, [currentQuestion?.questionId]);

  async function handleStart(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    setIsSubmitting(true);

    try {
      const payload = Object.fromEntries(
        Object.entries(metadata).filter(([, value]) => value.trim() !== "")
      );

      const result = await api.startSession({
        surveyId,
        shareLinkId: setup?.shareLinkId,
        clientMetadata: payload
      });

      setSessionId(result.sessionId);
      setCurrentQuestion(result.firstQuestion);
      setPhase("testing");
    } catch (startError) {
      setError(readErrorMessage(startError));
    } finally {
      setIsSubmitting(false);
    }
  }

  async function handleSubmitAnswer(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!currentQuestion) {
      return;
    }

    setError("");
    setIsSubmitting(true);

    try {
      const result = await api.submitAnswer({
        sessionId,
        questionId: currentQuestion.questionId,
        answerId: currentAnswerId || undefined,
        answerIds: currentAnswerIds.length > 0 ? currentAnswerIds : undefined,
        rawText: rawText.trim() || undefined
      });

      if (result.isFinished) {
        setReportDelivery(result.reportDelivery ?? null);
        setPhase("finished");
        return;
      }

      if (result.nextQuestion) {
        setCurrentQuestion(result.nextQuestion);
      }
    } catch (answerError) {
      setError(readErrorMessage(answerError));
    } finally {
      setIsSubmitting(false);
    }
  }

  const heroTitle = setup?.title ?? "ПрофДНК";
  const heroDescription = setup?.description ?? "Пожалуйста, заполните короткую анкету и пройдите диагностику до конца.";

  return (
    <main className="candidate-page">
      <BrandHomeLink compact className="brand-floating" />
      <section className="candidate-page__intro">
        <Badge>Public flow</Badge>
        <h1>{heroTitle}</h1>
        <p>{heroDescription}</p>
        <span>{setup?.intro ?? "После завершения система автоматически сформирует отчет и отправит его на указанный email."}</span>
      </section>

      {phase === "start" ? (
        <Card className="candidate-card">
          <div className="candidate-card__header">
            <div>
              <p className="eyebrow">Step 1</p>
              <h2>Анкета перед началом</h2>
            </div>
            <Badge>{surveyId.slice(0, 8)}</Badge>
          </div>

          <form className="form-grid" onSubmit={handleStart}>
            {fields.map((field) => (
              <Field
                key={field.key}
                label={field.label}
                hint={field.required ? "Обязательное поле" : "Можно оставить пустым"}
              >
                <Input
                  type={field.kind}
                  required={field.required}
                  value={metadata[field.key] ?? ""}
                  onChange={(event) => setMetadata((current) => ({
                    ...current,
                    [field.key]: event.target.value
                  }))}
                  placeholder={field.placeholder}
                />
              </Field>
            ))}

            {error ? <p className="feedback feedback--error">{error}</p> : null}

            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Запускаем тест..." : "Начать прохождение"}
            </Button>
          </form>
        </Card>
      ) : null}

      {phase === "testing" && currentQuestion ? (
        <Card className="candidate-card">
          <div className="candidate-card__header">
            <div>
              <p className="eyebrow">Step 2</p>
              <h2>{currentQuestion.text}</h2>
            </div>
            <Badge>{currentQuestion.type}</Badge>
          </div>

          {currentQuestion.helperText ? <p className="muted">{currentQuestion.helperText}</p> : null}

          <form className="stack" onSubmit={handleSubmitAnswer}>
            {currentQuestion.type === "text" ? (
              <Field label="Ваш ответ">
                <TextArea
                  rows={6}
                  value={rawText}
                  onChange={(event) => setRawText(event.target.value)}
                  placeholder="Опишите ваш ответ подробнее"
                  required
                />
              </Field>
            ) : currentQuestion.type === "multiple_choice" ? (
              <div className="answer-grid">
                {currentQuestion.answers.map((answer) => (
                  <label key={answer.answerId} className="answer-option">
                    <input
                      type="checkbox"
                      checked={currentAnswerIds.includes(answer.answerId)}
                      onChange={(event) => {
                        if (event.target.checked) {
                          setCurrentAnswerIds((current) => [...current, answer.answerId]);
                        } else {
                          setCurrentAnswerIds((current) => current.filter((item) => item !== answer.answerId));
                        }
                      }}
                    />
                    <span>{answer.text}</span>
                  </label>
                ))}
              </div>
            ) : (
              <div className="answer-grid">
                {currentQuestion.answers.map((answer) => (
                  <label key={answer.answerId} className="answer-option">
                    <input
                      type="radio"
                      name="single-answer"
                      checked={currentAnswerId === answer.answerId}
                      onChange={() => setCurrentAnswerId(answer.answerId)}
                    />
                    <span>{answer.text}</span>
                  </label>
                ))}
              </div>
            )}

            {error ? <p className="feedback feedback--error">{error}</p> : null}

            <Button
              type="submit"
              disabled={
                isSubmitting
                || (currentQuestion.type === "text" && rawText.trim().length === 0)
                || (currentQuestion.type === "multiple_choice" && currentAnswerIds.length === 0)
                || (currentQuestion.type !== "text" && currentQuestion.type !== "multiple_choice" && !currentAnswerId)
              }
            >
              {isSubmitting ? "Сохраняем..." : "Сохранить ответ и продолжить"}
            </Button>
          </form>
        </Card>
      ) : null}

      {phase === "finished" ? (
        <Card className="candidate-card candidate-card--success">
          <div className="candidate-card__header">
            <div>
              <p className="eyebrow">Step 3</p>
              <h2>Отчет уже в пути</h2>
            </div>
            <Badge>Done</Badge>
          </div>
          <p>{reportDelivery?.status === "sent"
            ? `Отчет сформирован и отправлен на ${reportDelivery.email}. Эту сессию повторно пройти уже нельзя.`
            : "Ответы сохранены, но отправка отчета не удалась. Психолог сможет переотправить его из кабинета."}
          </p>
          <p className="muted">
            {reportDelivery?.status === "sent"
              ? "Письмо приходит автоматически на финише после обращения к аналитике."
              : "Если письмо не дошло, проверьте адрес email и свяжитесь с психологом."}
          </p>
          <Button type="button" disabled>
            {reportDelivery?.status === "sent" ? "Отчет отправлен на почту" : "Сессия закрыта"}
          </Button>
        </Card>
      ) : null}
    </main>
  );
}
