import { AppShell, Text } from "@mantine/core";
import { Route, Routes, useParams } from "react-router";
import { ContextView } from "./components/ContextView";
import { Header } from "./components/Header";
import { Sidebar } from "./components/Sidebar";
import { CurrentContexts } from "./pages/CurrentContexts";

function Home() {
  return <Text c="dimmed">Select a context in the sidebar.</Text>;
}

// ContextPage reads the context id from the route: /contexts/<id>.
function ContextPage() {
  const id = useParams()["*"] ?? "";
  return <ContextView key={id} id={id} />;
}

export function App() {
  return (
    <AppShell header={{ height: 56 }} navbar={{ width: 300, breakpoint: "sm" }} padding="lg">
      <AppShell.Header>
        <Header />
      </AppShell.Header>
      <AppShell.Navbar>
        <Sidebar />
      </AppShell.Navbar>
      <AppShell.Main>
        <Routes>
          <Route path="/contexts/*" element={<ContextPage />} />
          <Route path="/current" element={<CurrentContexts />} />
          <Route path="*" element={<Home />} />
        </Routes>
      </AppShell.Main>
    </AppShell>
  );
}
