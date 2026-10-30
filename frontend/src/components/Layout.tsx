import { NavLink, Outlet } from "react-router-dom";

export function Layout() {
  return (
    <>
      <header className="header">
        <span className="brand">User Management</span>
        <nav>
          <NavLink to="/" end>
            Dashboard
          </NavLink>
          <NavLink to="/users">Users</NavLink>
        </nav>
      </header>
      <main className="main">
        <Outlet />
      </main>
    </>
  );
}
