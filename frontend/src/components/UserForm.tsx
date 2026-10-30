import { useState, type FormEvent } from "react";
import { ApiError } from "../api/client";
import type { UserInput } from "../types/user";
import { validateUser, type FormErrors } from "../validation/userValidation";
import { useNotify } from "./NotificationProvider";

interface Props {
  initial?: UserInput;
  submitLabel: string;
  onSubmit: (input: UserInput) => Promise<void>;
}

const EMPTY: UserInput = { firstName: "", lastName: "", email: "", phone: "", department: "", designation: "" };

const FIELDS: { key: keyof UserInput; label: string; type?: string }[] = [
  { key: "firstName", label: "First Name" },
  { key: "lastName", label: "Last Name" },
  { key: "email", label: "Email Address", type: "email" },
  { key: "phone", label: "Phone Number", type: "tel" },
  { key: "department", label: "Department" },
  { key: "designation", label: "Designation" },
];

export function UserForm({ initial = EMPTY, submitLabel, onSubmit }: Props) {
  const [values, setValues] = useState<UserInput>(initial);
  const [errors, setErrors] = useState<FormErrors>({});
  const [busy, setBusy] = useState(false);
  const notify = useNotify();

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    const found = validateUser(values);
    setErrors(found);
    if (Object.keys(found).length > 0) return;
    setBusy(true);
    try {
      await onSubmit({ ...values, email: values.email.trim().toLowerCase() });
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setErrors({ email: "A user with this email already exists." });
      } else if (err instanceof ApiError && err.status === 400 && err.details) {
        setErrors(err.details as FormErrors);
      } else {
        notify.error(err instanceof Error ? err.message : "Something went wrong.");
      }
    } finally {
      setBusy(false);
    }
  };

  return (
    <form className="user-form" onSubmit={submit} noValidate>
      {FIELDS.map(({ key, label, type }) => (
        <div className="field" key={key}>
          <label htmlFor={key}>{label}</label>
          <input
            id={key}
            type={type ?? "text"}
            value={values[key]}
            aria-invalid={errors[key] ? true : undefined}
            onChange={(e) => setValues({ ...values, [key]: e.target.value })}
          />
          {errors[key] && (
            <span className="field-error" role="alert">
              {errors[key]}
            </span>
          )}
        </div>
      ))}
      <button type="submit" disabled={busy}>
        {submitLabel}
      </button>
    </form>
  );
}
