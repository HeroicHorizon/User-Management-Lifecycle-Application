import { Link } from "react-router-dom";
import type { User } from "../types/user";

interface Props {
  users: User[];
  onToggleStatus: (u: User) => void;
  onDelete: (u: User) => void;
}

const fmt = (iso: string) => new Date(iso).toLocaleString();

export function UserTable({ users, onToggleStatus, onDelete }: Props) {
  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>User ID</th>
            <th>First Name</th>
            <th>Last Name</th>
            <th>Email Address</th>
            <th>Phone Number</th>
            <th>Department</th>
            <th>Designation</th>
            <th>Status</th>
            <th>Created Date</th>
            <th>Updated Date</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          {users.map((u) => (
            <tr key={u.id} data-testid="user-row">
              <td>{u.id}</td>
              <td>{u.firstName}</td>
              <td>{u.lastName}</td>
              <td>{u.email}</td>
              <td>{u.phone}</td>
              <td>{u.department}</td>
              <td>{u.designation}</td>
              <td>
                <span className={`badge badge-${u.status.toLowerCase()}`}>{u.status}</span>
              </td>
              <td>{fmt(u.createdAt)}</td>
              <td>{fmt(u.updatedAt)}</td>
              <td className="row-actions">
                <Link to={`/users/${u.id}`}>View</Link>
                <Link to={`/users/${u.id}/edit`}>Edit</Link>
                <button type="button" onClick={() => onToggleStatus(u)}>
                  {u.status === "ACTIVE" ? "Deactivate" : "Activate"}
                </button>
                <button type="button" className="danger" onClick={() => onDelete(u)}>
                  Delete
                </button>
              </td>
            </tr>
          ))}
          {users.length === 0 && (
            <tr>
              <td colSpan={11} className="empty">
                No users found.
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  );
}
