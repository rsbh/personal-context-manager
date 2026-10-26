import { Text } from "@raystack/apsara";
import { Outlet, useParams, type RouteObject } from "react-router";
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

// Layout is the app shell: header, sidebar and the routed page.
function Layout() {
  return (
    <div className={classes.shell}>
      <Header />
      <div className={classes.body}>
        <Sidebar />
        <main className={classes.main}>
          <Outlet />
        </main>
      </div>
    </div>
  );
}

// routes are the app routes for createBrowserRouter.
export const routes: RouteObject[] = [
  {
    element: <Layout />,
    children: [
      { path: "/contexts/*", element: <ContextPage /> },
      { path: "/current", element: <CurrentContexts /> },
      { path: "*", element: <Home /> },
    ],
  },
];
