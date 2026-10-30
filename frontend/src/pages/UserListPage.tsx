import { useCallback, useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { deleteUser, listUsers, setUserStatus } from "../api/users";
import { ConfirmDialog } from "../components/ConfirmDialog";
import { useNotify } from "../components/NotificationProvider";
import { Pagination } from "../components/Pagination";
import { SearchBar } from "../components/SearchBar";
import { StatusFilter } from "../components/StatusFilter";
import { UserTable } from "../components/UserTable";
import type { ListParams, User, UserList } from "../types/user";

export const PAGE_SIZE = 10;

export function UserListPage() {
  const notify = useNotify();
  const [params, setParams] = useState<ListParams>({ page: 1, pageSize: PAGE_SIZE, search: "", status: "" });
  const [data, setData] = useState<UserList>({ items: [], page: 1, pageSize: PAGE_SIZE, total: 0 });
  const [pendingDelete, setPendingDelete] = useState<User | null>(null);
  const [reload, setReload] = useState(0);

  useEffect(() => {
    let stale = false;
    listUsers(params)
      .then((r) => !stale && setData(r))
      .catch((e: Error) => !stale && notify.error(e.message));
    return () => {
      stale = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [params, reload]);

  const refresh = useCallback(() => setReload((n) => n + 1), []);

  const toggle = async (u: User) => {
    try {
      const next = u.status === "ACTIVE" ? "INACTIVE" : "ACTIVE";
      await setUserStatus(u.id, next);
      notify.success(`User ${next === "ACTIVE" ? "activated" : "deactivated"}.`);
      refresh();
    } catch (e) {
      notify.error((e as Error).message);
    }
  };

  const confirmDelete = async () => {
    const u = pendingDelete;
    setPendingDelete(null);
    if (!u) return;
    try {
      await deleteUser(u.id);
      notify.success("User deleted.");
      refresh();
    } catch (e) {
      notify.error((e as Error).message);
    }
  };

  return (
    <section>
      <div className="toolbar">
        <h1>User Management</h1>
        <Link to="/users/new" className="button">
          Add User
        </Link>
      </div>
      <div className="filters">
        <SearchBar onSearch={(search) => setParams((p) => ({ ...p, search, page: 1 }))} />
        <StatusFilter value={params.status} onChange={(status) => setParams((p) => ({ ...p, status, page: 1 }))} />
      </div>
      <UserTable users={data.items} onToggleStatus={toggle} onDelete={setPendingDelete} />
      <Pagination
        page={params.page}
        pageSize={data.pageSize}
        total={data.total}
        onPage={(page) => setParams((p) => ({ ...p, page }))}
      />
      {pendingDelete && (
        <ConfirmDialog
          message={`Delete ${pendingDelete.firstName} ${pendingDelete.lastName}? This cannot be undone.`}
          onConfirm={confirmDelete}
          onCancel={() => setPendingDelete(null)}
        />
      )}
    </section>
  );
}
