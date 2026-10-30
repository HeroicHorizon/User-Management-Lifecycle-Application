import { useEffect, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { getUser } from "../api/users";
import { useNotify } from "../components/NotificationProvider";
import type { User } from "../types/user";

export function UserDetailsPage() {
  const { id = "" } = useParams();
  const notify = useNotify();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    getUser(id)
      .then(setUser)
      .catch((e: Error) => notify.error(e.message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  if (!user) return <p>Loading…</p>;
  const rows: [string, string][] = [
    ["User ID", user.id],
    ["First Name", user.firstName],
    ["Last Name", user.lastName],
    ["Email Address", user.email],
    ["Phone Number", user.phone],
    ["Department", user.department],
    ["Designation", user.designation],
    ["Status", user.status],
    ["Created Date", new Date(user.createdAt).toLocaleString()],
    ["Updated Date", new Date(user.updatedAt).toLocaleString()],
  ];
  return (
    <section>
      <h1>User Details</h1>
      <dl className="details">
        {rows.map(([k, v]) => (
          <div key={k}>
            <dt>{k}</dt>
            <dd>{v}</dd>
          </div>
        ))}
      </dl>
      <Link to={`/users/${user.id}/edit`}>Edit</Link> <Link to="/users">Back to list</Link>
    </section>
  );
}
