import { Text } from "@raystack/apsara";
import { Route, Routes, useParams } from "react-router";
import classes from "./App.module.css";
import { ContextView } from "./components/ContextView";
import { Header } from "./components/Header";
import { Sidebar } from "./components/Sidebar";
import { CurrentContexts } from "./pages/CurrentContexts";

function Home() {
  return <Text variant="secondary">Select a context in the sidebar.</Text>;
}

// ContextPage reads the context id from the route: /contexts/<id>.
function ContextPage() {
  const id = useParams()["*"] ?? "";
  return <ContextView key={id} id={id} />;
}

const contextPage = <ContextPage />;
const currentContexts = <CurrentContexts />;
const home = <Home />;

export function App() {
  return (
    <div className={classes.shell}>
      <Header />
      <div className={classes.body}>
        <Sidebar />
        <main className={classes.main}>
          <Routes>
            <Route path="/contexts/*" element={contextPage} />
            <Route path="/current" element={currentContexts} />
            <Route path="*" element={home} />
          </Routes>
        </main>
      </div>
    </div>
  );
}
