import type { UserInput } from "../types/user";

export type FormErrors = Partial<Record<keyof UserInput, string>>;

const EMAIL_RE = /^[A-Za-z0-9._%+-]+@[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)*\.[A-Za-z]{2,}$/;
const PHONE_RE = /^\+?[0-9][0-9 -]{5,19}$/;

const REQUIRED: (keyof UserInput)[] = ["firstName", "lastName", "email", "department", "designation"];

export function validateUser(input: UserInput): FormErrors {
  const errors: FormErrors = {};
  for (const key of REQUIRED) {
    if (!input[key].trim()) errors[key] = "This field is required.";
  }
  const email = input.email.trim();
  if (email && !EMAIL_RE.test(email)) errors.email = "Enter a valid email address.";
  const phone = input.phone.trim();
  if (phone && !PHONE_RE.test(phone)) errors.phone = "Enter a valid phone number.";
  return errors;
}
