import { useState, type FormEvent } from "react";

export function SearchBar({ onSearch }: { onSearch: (term: string) => void }) {
  const [term, setTerm] = useState("");
  const submit = (e: FormEvent) => {
    e.preventDefault();
    onSearch(term);
  };
  return (
    <form className="search" onSubmit={submit} role="search">
      <input
        type="search"
        aria-label="Search users"
        placeholder="Search by name or email"
        value={term}
        onChange={(e) => setTerm(e.target.value)}
      />
      <button type="submit">Search</button>
    </form>
  );
}
