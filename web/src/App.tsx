import { AppShell, Text } from "@mantine/core";
import { Route, Routes } from "react-router";
import { Header } from "./components/Header";
import { Sidebar } from "./components/Sidebar";

function Home() {
  return <Text c="dimmed">Select a context in the sidebar.</Text>;
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
          <Route path="*" element={<Home />} />
        </Routes>
      </AppShell.Main>
    </AppShell>
  );
}
