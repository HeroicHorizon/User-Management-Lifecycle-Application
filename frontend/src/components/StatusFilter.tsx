import type { UserStatus } from "../types/user";

interface Props {
  value: UserStatus | "";
  onChange: (v: UserStatus | "") => void;
}

export function StatusFilter({ value, onChange }: Props) {
  return (
    <select aria-label="Filter by status" value={value} onChange={(e) => onChange(e.target.value as UserStatus | "")}>
      <option value="">All statuses</option>
      <option value="ACTIVE">ACTIVE</option>
      <option value="INACTIVE">INACTIVE</option>
    </select>
  );
}
