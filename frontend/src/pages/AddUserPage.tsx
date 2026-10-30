import { useNavigate } from "react-router-dom";
import { createUser } from "../api/users";
import { useNotify } from "../components/NotificationProvider";
import { UserForm } from "../components/UserForm";

export function AddUserPage() {
  const navigate = useNavigate();
  const notify = useNotify();
  return (
    <section>
      <h1>Add User</h1>
      <UserForm
        submitLabel="Create User"
        onSubmit={async (input) => {
          await createUser(input);
          notify.success("User created successfully.");
          navigate("/users");
        }}
      />
    </section>
  );
}
