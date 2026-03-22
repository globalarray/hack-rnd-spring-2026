import { Link } from "react-router-dom";

import { BrandAttribution, BrandHomeLink, BRAND_LINKS } from "../components/brand";
import { Badge, Card } from "../components/ui";

export function HomePage() {
  return (
    <main className="home-page">
      <header className="home-page__header">
        <BrandHomeLink />
        <div className="home-page__header-links">
          <a href={BRAND_LINKS.benzo} target="_blank" rel="noreferrer">
            benzo.cloud
          </a>
          <a href={BRAND_LINKS.telegram} target="_blank" rel="noreferrer">
            Связаться
          </a>
        </div>
      </header>

      <section className="home-hero">
        <div className="home-hero__copy">
          <Badge>Platform overview</Badge>
          <h1>Платформа для психологов, администраторов и управляемого прохождения диагностик.</h1>
          <p>
            ПрофДНК помогает собирать методики в визуальном конструкторе, приглашать специалистов по ссылке,
            запускать неограниченное число публичных сессий и доводить каждый тест до персонального отчета.
          </p>
          <BrandAttribution />
          <div className="home-hero__actions">
            <Link className="home-cta" to="/login">
              Открыть кабинет
            </Link>
            <a className="home-cta home-cta--ghost" href={BRAND_LINKS.telegram} target="_blank" rel="noreferrer">
              Связаться в Telegram
            </a>
          </div>
        </div>

        <Card className="home-hero__spotlight">
          <strong>Что умеет продукт</strong>
          <ul>
            <li>Единый вход для администратора и психолога.</li>
            <li>Invitation-based onboarding для психологов через `/invitations/{uuid}`.</li>
            <li>Drag-and-drop конструктор тестов с логикой переходов между вопросами.</li>
            <li>Публичные ссылки на прохождение с настраиваемой metadata перед стартом.</li>
            <li>Email-отчеты для клиента и отдельный отчет для психолога.</li>
          </ul>
        </Card>
      </section>

      <section className="home-grid">
        <Card>
          <h2>Для администраторов</h2>
          <p>
            Управление доступом, создание профилей психологов, выдача invitation-ссылок, блокировка и контроль
            жизненного цикла аккаунтов в одном окне.
          </p>
        </Card>
        <Card>
          <h2>Для психологов</h2>
          <p>
            Конструктор методик, список тестов, результаты прохождений, отправка отчетов клиенту и личный
            профессиональный разбор по каждой завершенной сессии.
          </p>
        </Card>
        <Card>
          <h2>Для кандидатов</h2>
          <p>
            Простой flow: переход по ссылке, заполнение metadata, прохождение вопросов и автоматическая отправка
            отчета на email после завершения сессии.
          </p>
        </Card>
      </section>
    </main>
  );
}
