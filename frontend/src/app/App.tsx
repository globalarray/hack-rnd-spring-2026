import { BrowserRouter, Navigate, Outlet, Route, Routes } from "react-router-dom";

import { LoadingScreen } from "../components/ui";
import { PUBLIC_APP_URL } from "../lib/app-config";
import { AdminDashboard } from "../pages/AdminDashboard";
import { CandidateFlowPage } from "../pages/CandidateFlowPage";
import { HomePage } from "../pages/HomePage";
import { InvitationPage } from "../pages/InvitationPage";
import { LoginPage } from "../pages/LoginPage";
import { PsychologistDashboard } from "../pages/PsychologistDashboard";
import { AuthProvider, useAuth } from "./auth";
import { AppErrorBoundary } from "./error-boundary";

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route element={<GuestOnlyRoute />}>
        <Route path="/login" element={<LoginPage />} />
      </Route>
      <Route path="/invitations/:token" element={<InvitationPage />} />
      <Route path="/tests/:surveyId/start" element={<CandidateFlowPage />} />
      <Route element={<ProtectedRoute />}>
        <Route path="/admin" element={<AdminDashboard />} />
        <Route path="/psychologist" element={<PsychologistDashboard />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}

function ProtectedRoute() {
  const { session, isBooting } = useAuth();

  if (isBooting) {
    return <LoadingScreen title="Подготавливаем кабинет" description="Еще несколько секунд, и все разделы будут доступны." />;
  }

  if (!session) {
    return <Navigate to="/login" replace />;
  }

  return <Outlet />;
}

function GuestOnlyRoute() {
  const { session, isBooting } = useAuth();

  if (isBooting) {
    return <LoadingScreen title="Проверяем сессию" description="Если вы уже входили в систему, сразу откроем нужную панель." />;
  }

  if (session) {
    return <Navigate to={session.profile.role === "admin" ? "/admin" : "/psychologist"} replace />;
  }

  return <Outlet />;
}

export function App() {
  if (typeof window !== "undefined") {
    try {
      const canonical = new URL(PUBLIC_APP_URL);
      const current = new URL(window.location.href);
      if (current.host !== canonical.host || current.protocol !== canonical.protocol) {
        window.location.replace(`${canonical.origin}${current.pathname}${current.search}${current.hash}`);
        return <LoadingScreen title="Перенаправляем на основной домен" description="Открываем приложение на hack.benzo.cloud:3000." />;
      }
    } catch {
      // noop: keep the app available even if the configured URL is malformed
    }
  }

  return (
    <AuthProvider>
      <BrowserRouter>
        <AppErrorBoundary>
          <AppRoutes />
        </AppErrorBoundary>
      </BrowserRouter>
    </AuthProvider>
  );
}
