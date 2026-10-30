import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { getUser, updateUser } from "../api/users";
import { useNotify } from "../components/NotificationProvider";
import { UserForm } from "../components/UserForm";
import type { User } from "../types/user";

export function EditUserPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const notify = useNotify();
  const [user, setUser] = useState<User | null>(null);

  useEffect(() => {
    getUser(id)
      .then(setUser)
      .catch((e: Error) => notify.error(e.message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  if (!user) return <p>Loading…</p>;
  return (
    <section>
      <h1>Edit User</h1>
      <UserForm
        initial={user}
        submitLabel="Save Changes"
        onSubmit={async (input) => {
          await updateUser(id, input);
          notify.success("User updated successfully.");
          navigate(`/users/${id}`);
        }}
      />
    </section>
  );
}
