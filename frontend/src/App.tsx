import { Route, Routes } from "react-router-dom";
import { Layout } from "./components/Layout";
import { AddUserPage } from "./pages/AddUserPage";
import { Dashboard } from "./pages/Dashboard";
import { EditUserPage } from "./pages/EditUserPage";
import { UserDetailsPage } from "./pages/UserDetailsPage";
import { UserListPage } from "./pages/UserListPage";

export function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/users" element={<UserListPage />} />
        <Route path="/users/new" element={<AddUserPage />} />
        <Route path="/users/:id" element={<UserDetailsPage />} />
        <Route path="/users/:id/edit" element={<EditUserPage />} />
      </Route>
    </Routes>
  );
}
