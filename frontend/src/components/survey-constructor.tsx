import { type DragEndEvent, DndContext, PointerSensor, closestCenter, useSensor, useSensors } from "@dnd-kit/core";
import { SortableContext, arrayMove, useSortable, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { useEffect, useMemo, useState } from "react";

import type { LogicAction, QuestionDraft, ShareLinkConfig, SurveyDraft } from "../lib/types";
import { buildQuestionTemplate, createId, cx, ensureQuestionOrder, normalizeFieldKey, toQuestionOptions } from "../lib/utils";
import { Badge, Button, Card, EmptyState, Field, GhostButton, Input, SectionTitle, Select, TextArea } from "./ui";

type SurveyConstructorProps = {
  draft: SurveyDraft;
  onChange: (nextDraft: SurveyDraft) => void;
  onSave: () => Promise<void>;
  isSaving: boolean;
  shareLinks: ShareLinkConfig[];
};

type SortableQuestionProps = {
  question: QuestionDraft;
  isSelected: boolean;
  onSelect: () => void;
  onDuplicate: () => void;
  onDelete: () => void;
};

type SortableAnswerProps = {
  id: string;
  label: string;
};

function SortableQuestionCard({ question, isSelected, onSelect, onDuplicate, onDelete }: SortableQuestionProps) {
  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id: question.id });

  return (
    <div
      ref={setNodeRef}
      role="button"
      tabIndex={0}
      className={cx("sortable-card", isSelected && "is-selected")}
      style={{
        transform: CSS.Transform.toString(transform),
        transition
      }}
      onClick={onSelect}
      onKeyDown={(event) => {
        if (event.key === "Enter" || event.key === " ") {
          onSelect();
        }
      }}
    >
      <div className="sortable-card__top">
        <span className="drag-handle" {...attributes} {...listeners}>
          drag
        </span>
        <Badge>{question.type}</Badge>
      </div>
      <strong>{question.orderNum}. {question.text || "Новый вопрос"}</strong>
      <p>{question.helperText || "Без пояснения"}</p>
      <div className="sortable-card__actions">
        <GhostButton type="button" onClick={(event) => { event.stopPropagation(); onDuplicate(); }}>
          Дублировать
        </GhostButton>
        <GhostButton type="button" onClick={(event) => { event.stopPropagation(); onDelete(); }}>
          Удалить
        </GhostButton>
      </div>
    </div>
  );
}

function SortableAnswerRow({ id, label }: SortableAnswerProps) {
  const { attributes, listeners, setNodeRef, transform, transition } = useSortable({ id });

  return (
    <div
      ref={setNodeRef}
      className="answer-chip"
      style={{
        transform: CSS.Transform.toString(transform),
        transition
      }}
    >
      <span className="drag-handle" {...attributes} {...listeners}>
        drag
      </span>
      <span>{label}</span>
    </div>
  );
}

export function SurveyConstructor({ draft, onChange, onSave, isSaving, shareLinks }: SurveyConstructorProps) {
  const [selectedQuestionId, setSelectedQuestionId] = useState(draft.questions[0]?.id ?? "");
  const [newFieldLabel, setNewFieldLabel] = useState("");
  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 6 } }));

  useEffect(() => {
    if (!draft.questions.find((question) => question.id === selectedQuestionId)) {
      setSelectedQuestionId(draft.questions[0]?.id ?? "");
    }
  }, [draft.questions, selectedQuestionId]);

  const selectedQuestion = useMemo(
    () => draft.questions.find((question) => question.id === selectedQuestionId) ?? null,
    [draft.questions, selectedQuestionId]
  );
  const questionOptions = toQuestionOptions(draft.questions);

  function updateDraft(mutator: (current: SurveyDraft) => SurveyDraft) {
    onChange(mutator(draft));
  }

  function handleQuestionDrag(event: DragEndEvent) {
    const { active, over } = event;
    if (!over || active.id === over.id) {
      return;
    }

    const oldIndex = draft.questions.findIndex((question) => question.id === active.id);
    const newIndex = draft.questions.findIndex((question) => question.id === over.id);
    if (oldIndex < 0 || newIndex < 0) {
      return;
    }

    updateDraft((current) => ({
      ...current,
      questions: ensureQuestionOrder(arrayMove(current.questions, oldIndex, newIndex))
    }));
  }

  function handleAnswerDrag(event: DragEndEvent) {
    if (!selectedQuestion) {
      return;
    }

    const { active, over } = event;
    if (!over || active.id === over.id) {
      return;
    }

    const oldIndex = selectedQuestion.answers.findIndex((answer) => answer.id === active.id);
    const newIndex = selectedQuestion.answers.findIndex((answer) => answer.id === over.id);
    if (oldIndex < 0 || newIndex < 0) {
      return;
    }

    updateQuestion(selectedQuestion.id, {
      answers: arrayMove(selectedQuestion.answers, oldIndex, newIndex)
    });
  }

  function addQuestion(type: QuestionDraft["type"]) {
    const question = buildQuestionTemplate(type, draft.questions.length + 1);
    updateDraft((current) => ({
      ...current,
      questions: [...current.questions, question]
    }));
    setSelectedQuestionId(question.id);
  }

  function duplicateQuestion(questionId: string) {
    const source = draft.questions.find((question) => question.id === questionId);
    if (!source) {
      return;
    }

    const answerIdMap = new Map(source.answers.map((answer) => [answer.id, createId()]));
    const duplicated: QuestionDraft = {
      ...source,
      id: createId(),
      answers: source.answers.map((answer) => ({ ...answer, id: answerIdMap.get(answer.id) ?? answer.id })),
      logicRules: source.logicRules.map((rule) => ({
        ...rule,
        answerId: answerIdMap.get(rule.answerId) ?? rule.answerId
      }))
    };

    updateDraft((current) => {
      const index = current.questions.findIndex((question) => question.id === questionId);
      const next = [...current.questions];
      next.splice(index + 1, 0, duplicated);
      return {
        ...current,
        questions: ensureQuestionOrder(next)
      };
    });
    setSelectedQuestionId(duplicated.id);
  }

  function deleteQuestion(questionId: string) {
    updateDraft((current) => ({
      ...current,
      questions: ensureQuestionOrder(current.questions.filter((question) => question.id !== questionId))
    }));
  }

  function updateQuestion(questionId: string, patch: Partial<QuestionDraft>) {
    updateDraft((current) => ({
      ...current,
      questions: current.questions.map((question) => (
        question.id === questionId ? { ...question, ...patch } : question
      ))
    }));
  }

  function updateAnswer(questionId: string, answerId: string, patch: Partial<QuestionDraft["answers"][number]>) {
    const question = draft.questions.find((item) => item.id === questionId);
    if (!question) {
      return;
    }

    updateQuestion(questionId, {
      answers: question.answers.map((answer) => (
        answer.id === answerId ? { ...answer, ...patch } : answer
      ))
    });
  }

  function updateLogicRule(questionId: string, answerId: string, patch: { action: LogicAction; nextQuestionId?: string }) {
    const question = draft.questions.find((item) => item.id === questionId);
    if (!question) {
      return;
    }

    const existing = question.logicRules.find((rule) => rule.answerId === answerId);
    const nextRules = existing
      ? question.logicRules.map((rule) => rule.answerId === answerId ? { ...rule, ...patch } : rule)
      : [...question.logicRules, { answerId, ...patch }];

    updateQuestion(questionId, { logicRules: nextRules });
  }

  function addAnswer() {
    if (!selectedQuestion) {
      return;
    }

    updateQuestion(selectedQuestion.id, {
      answers: [
        ...selectedQuestion.answers,
        {
          id: createId(),
          text: "Новый вариант",
          weight: selectedQuestion.answers.length + 1
        }
      ]
    });
  }

  function removeAnswer(answerId: string) {
    if (!selectedQuestion) {
      return;
    }

    updateQuestion(selectedQuestion.id, {
      answers: selectedQuestion.answers.filter((answer) => answer.id !== answerId),
      logicRules: selectedQuestion.logicRules.filter((rule) => rule.answerId !== answerId)
    });
  }

  function addCustomField() {
    const label = newFieldLabel.trim();
    if (!label) {
      return;
    }

    updateDraft((current) => ({
      ...current,
      settings: {
        ...current.settings,
        startForm: {
          ...current.settings.startForm,
          fields: [
            ...current.settings.startForm.fields,
            {
              key: normalizeFieldKey(label),
              label,
              required: false,
              kind: "text",
              placeholder: label
            }
          ]
        }
      }
    }));
    setNewFieldLabel("");
  }

  function updateField(index: number, patch: Partial<SurveyDraft["settings"]["startForm"]["fields"][number]>) {
    updateDraft((current) => ({
      ...current,
      settings: {
        ...current.settings,
        startForm: {
          ...current.settings.startForm,
          fields: current.settings.startForm.fields.map((field, currentIndex) => (
            currentIndex === index ? { ...field, ...patch } : field
          ))
        }
      }
    }));
  }

  function removeField(index: number) {
    if (index < 2) {
      return;
    }

    updateDraft((current) => ({
      ...current,
      settings: {
        ...current.settings,
        startForm: {
          ...current.settings.startForm,
          fields: current.settings.startForm.fields.filter((_, currentIndex) => currentIndex !== index)
        }
      }
    }));
  }

  return (
    <div className="stack constructor-stack">
      <Card className="stack">
        <SectionTitle
          title="Паспорт методики"
          description="Общая информация о тесте, вступительный текст и набор полей, которые кандидат увидит до старта."
          action={
            <Button type="button" onClick={() => void onSave()} disabled={isSaving}>
              {isSaving ? "Сохраняем..." : "Сохранить тест"}
            </Button>
          }
        />

        <div className="form-grid form-grid--wide">
          <Field label="Название теста">
            <Input
              value={draft.title}
              onChange={(event) => updateDraft((current) => ({ ...current, title: event.target.value }))}
              placeholder="Например: Диагностика интересов"
            />
          </Field>
          <Field label="Описание">
            <TextArea
              rows={4}
              value={draft.description}
              onChange={(event) => updateDraft((current) => ({ ...current, description: event.target.value }))}
              placeholder="Какую задачу решает методика и для какой аудитории она подходит?"
            />
          </Field>
          <Field label="Лимит времени, сек">
            <Input
              type="number"
              min={60}
              value={draft.settings.limits.timeLimitSec}
              onChange={(event) => updateDraft((current) => ({
                ...current,
                settings: {
                  ...current.settings,
                  limits: {
                    timeLimitSec: Number(event.target.value)
                  }
                }
              }))}
            />
          </Field>
          <Field label="Текст перед стартом">
            <TextArea
              rows={4}
              value={draft.settings.startForm.intro}
              onChange={(event) => updateDraft((current) => ({
                ...current,
                settings: {
                  ...current.settings,
                  startForm: {
                    ...current.settings.startForm,
                    intro: event.target.value
                  }
                }
              }))}
            />
          </Field>
          <Field label="Заголовок финального экрана">
            <Input
              value={draft.settings.startForm.completionTitle}
              onChange={(event) => updateDraft((current) => ({
                ...current,
                settings: {
                  ...current.settings,
                  startForm: {
                    ...current.settings.startForm,
                    completionTitle: event.target.value
                  }
                }
              }))}
            />
          </Field>
          <Field label="Текст после завершения">
            <TextArea
              rows={4}
              value={draft.settings.startForm.completionBody}
              onChange={(event) => updateDraft((current) => ({
                ...current,
                settings: {
                  ...current.settings,
                  startForm: {
                    ...current.settings.startForm,
                    completionBody: event.target.value
                  }
                }
              }))}
            />
          </Field>
        </div>

        <Card className="stack constructor-note">
          <strong>Поля стартовой анкеты</strong>
          <p>ФИО и email обязательны всегда. Дополнительные поля можно добавлять ниже и переиспользовать в share links.</p>
          <div className="field-list">
            {draft.settings.startForm.fields.map((field, index) => (
              <div key={`${field.key}-${index}`} className="field-list__item">
                <Input
                  value={field.label}
                  onChange={(event) => updateField(index, {
                    label: event.target.value,
                    key: normalizeFieldKey(event.target.value)
                  })}
                  disabled={index < 2}
                />
                <Select
                  value={field.kind}
                  onChange={(event) => updateField(index, { kind: event.target.value as typeof field.kind })}
                >
                  <option value="text">Текст</option>
                  <option value="email">Email</option>
                  <option value="number">Число</option>
                  <option value="tel">Телефон</option>
                </Select>
                <label className="checkbox-inline">
                  <input
                    type="checkbox"
                    checked={field.required}
                    onChange={(event) => updateField(index, { required: event.target.checked })}
                    disabled={index < 2}
                  />
                  Обязательное
                </label>
                <GhostButton type="button" onClick={() => removeField(index)} disabled={index < 2}>
                  Удалить
                </GhostButton>
              </div>
            ))}
          </div>

          <div className="inline-form">
            <Input
              value={newFieldLabel}
              onChange={(event) => setNewFieldLabel(event.target.value)}
              placeholder="Например: Школа, возраст, должность"
            />
            <Button type="button" onClick={addCustomField}>
              Добавить поле
            </Button>
          </div>
        </Card>
      </Card>

      <div className="constructor-layout">
        <Card className="stack constructor-panel">
          <SectionTitle
            title="Компоненты"
            description="Выбирайте тип вопроса и сразу добавляйте его в canvas."
          />
          <div className="tool-list">
            <Button type="button" onClick={() => addQuestion("single_choice")}>Single choice</Button>
            <Button type="button" onClick={() => addQuestion("multiple_choice")}>Multiple choice</Button>
            <Button type="button" onClick={() => addQuestion("scale")}>Scale</Button>
            <Button type="button" onClick={() => addQuestion("text")}>Text</Button>
          </div>

          <Card className="highlight-card">
            <strong>Preview share links</strong>
            <span>Для этого теста уже создано ссылок: {shareLinks.length}</span>
            {shareLinks[0] ? <p>{shareLinks[0].title}</p> : <p>Ссылки появятся после сохранения и настройки раздела Sessions.</p>}
          </Card>
        </Card>

        <Card className="stack constructor-canvas">
          <SectionTitle
            title="Canvas вопросов"
            description="Перетаскивайте карточки, меняйте порядок и выбирайте вопрос для детальной настройки."
          />

          {draft.questions.length === 0 ? (
            <EmptyState
              title="Пока нет вопросов"
              description="Начните с панели слева и добавьте первый блок методики."
            />
          ) : (
            <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleQuestionDrag}>
              <SortableContext items={draft.questions.map((question) => question.id)} strategy={verticalListSortingStrategy}>
                <div className="sortable-list">
                  {draft.questions.map((question) => (
                    <SortableQuestionCard
                      key={question.id}
                      question={question}
                      isSelected={selectedQuestionId === question.id}
                      onSelect={() => setSelectedQuestionId(question.id)}
                      onDuplicate={() => duplicateQuestion(question.id)}
                      onDelete={() => deleteQuestion(question.id)}
                    />
                  ))}
                </div>
              </SortableContext>
            </DndContext>
          )}
        </Card>

        <Card className="stack constructor-panel">
          <SectionTitle
            title="Инспектор"
            description="Здесь настраиваются текст вопроса, варианты ответов и логика переходов между блоками."
          />

          {!selectedQuestion ? (
            <EmptyState
              title="Выберите вопрос"
              description="Кликните по карточке в canvas, чтобы открыть настройки."
            />
          ) : (
            <>
              <Field label="Текст вопроса">
                <TextArea
                  rows={4}
                  value={selectedQuestion.text}
                  onChange={(event) => updateQuestion(selectedQuestion.id, { text: event.target.value })}
                />
              </Field>
              <Field label="Пояснение">
                <TextArea
                  rows={3}
                  value={selectedQuestion.helperText}
                  onChange={(event) => updateQuestion(selectedQuestion.id, { helperText: event.target.value })}
                />
              </Field>

              {selectedQuestion.type === "text" ? (
                <Card className="highlight-card">
                  <strong>Text question</strong>
                  <p>Для текстовых вопросов не требуются варианты ответов. Аналитика будет брать rawText напрямую.</p>
                </Card>
              ) : (
                <>
                  <div className="subsection-title">
                    <strong>Варианты ответа</strong>
                    <GhostButton type="button" onClick={addAnswer}>
                      Добавить ответ
                    </GhostButton>
                  </div>

                  <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleAnswerDrag}>
                    <SortableContext items={selectedQuestion.answers.map((answer) => answer.id)} strategy={verticalListSortingStrategy}>
                      <div className="answer-chip-list">
                        {selectedQuestion.answers.map((answer) => (
                          <SortableAnswerRow key={answer.id} id={answer.id} label={answer.text} />
                        ))}
                      </div>
                    </SortableContext>
                  </DndContext>

                  <div className="stack">
                    {selectedQuestion.answers.map((answer) => {
                      const rule = selectedQuestion.logicRules.find((item) => item.answerId === answer.id);
                      return (
                        <Card key={answer.id} className="stack answer-editor">
                          <Field label="Текст ответа">
                            <Input
                              value={answer.text}
                              onChange={(event) => updateAnswer(selectedQuestion.id, answer.id, { text: event.target.value })}
                            />
                          </Field>
                          <Field label="Вес">
                            <Input
                              type="number"
                              value={answer.weight}
                              onChange={(event) => updateAnswer(selectedQuestion.id, answer.id, { weight: Number(event.target.value) })}
                            />
                          </Field>
                          <Field label="Категория">
                            <Input
                              value={answer.categoryTag ?? ""}
                              onChange={(event) => updateAnswer(selectedQuestion.id, answer.id, { categoryTag: event.target.value })}
                              placeholder="people / analysis / systems"
                            />
                          </Field>
                          <Field label="Логика перехода">
                            <Select
                              value={rule?.action ?? "linear"}
                              onChange={(event) => updateLogicRule(selectedQuestion.id, answer.id, {
                                action: event.target.value as LogicAction,
                                nextQuestionId: rule?.nextQuestionId
                              })}
                            >
                              <option value="linear">Линейно</option>
                              <option value="jump">Прыжок на вопрос</option>
                              <option value="finish">Завершить тест</option>
                            </Select>
                          </Field>
                          {(rule?.action ?? "linear") === "jump" ? (
                            <Field label="Следующий вопрос">
                              <Select
                                value={rule?.nextQuestionId ?? ""}
                                onChange={(event) => updateLogicRule(selectedQuestion.id, answer.id, {
                                  action: "jump",
                                  nextQuestionId: event.target.value
                                })}
                              >
                                <option value="">Выберите вопрос</option>
                                {questionOptions
                                  .filter((option) => option.value !== selectedQuestion.id)
                                  .map((option) => (
                                    <option key={option.value} value={option.value}>
                                      {option.label}
                                    </option>
                                  ))}
                              </Select>
                            </Field>
                          ) : null}
                          <GhostButton type="button" onClick={() => removeAnswer(answer.id)}>
                            Удалить ответ
                          </GhostButton>
                        </Card>
                      );
                    })}
                  </div>
                </>
              )}
            </>
          )}
        </Card>
      </div>
    </div>
  );
}
