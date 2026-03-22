import { Component, type ErrorInfo, type ReactNode } from "react";

import { BrandHomeLink } from "../components/brand";
import { Button, Card, GhostButton } from "../components/ui";

type AppErrorBoundaryProps = {
  children: ReactNode;
};

type AppErrorBoundaryState = {
  hasError: boolean;
  message: string;
};

const APP_STORAGE_KEYS = [
  "profdnk.session.v1",
  "profdnk.workspace.v1",
  "profdnk.mock-db.v1"
];

export class AppErrorBoundary extends Component<AppErrorBoundaryProps, AppErrorBoundaryState> {
  state: AppErrorBoundaryState = {
    hasError: false,
    message: ""
  };

  static getDerivedStateFromError(error: Error): AppErrorBoundaryState {
    return {
      hasError: true,
      message: error.message || "Не удалось открыть страницу."
    };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error("AppErrorBoundary", error, info);
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      message: ""
    });
  };

  handleResetStorage = () => {
    if (typeof window !== "undefined") {
      APP_STORAGE_KEYS.forEach((key) => window.localStorage.removeItem(key));
      window.location.assign("/");
    }
  };

  render() {
    if (!this.state.hasError) {
      return this.props.children;
    }

    return (
      <main className="auth-page">
        <BrandHomeLink compact className="brand-floating" />
        <section className="auth-page__hero">
          <h1>Стартовая страница временно недоступна.</h1>
          <p>
            Мы перехватили ошибку фронтенда, чтобы сайт не оставался с пустым экраном. Чаще всего такое бывает
            из-за старых данных браузера после обновления интерфейса.
          </p>
        </section>
        <Card className="auth-card">
          <div className="stack">
            <p className="feedback feedback--error">{this.state.message}</p>
            <Button type="button" onClick={this.handleRetry}>
              Попробовать еще раз
            </Button>
            <GhostButton type="button" onClick={this.handleResetStorage}>
              Очистить локальные данные и открыть главную
            </GhostButton>
          </div>
        </Card>
      </main>
    );
  }
}
