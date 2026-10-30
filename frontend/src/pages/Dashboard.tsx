import { useEffect, useState } from "react";
import { getStats } from "../api/users";
import { useNotify } from "../components/NotificationProvider";
import type { UserStats } from "../types/user";

export function Dashboard() {
  const notify = useNotify();
  const [stats, setStats] = useState<UserStats | null>(null);

  useEffect(() => {
    getStats()
      .then(setStats)
      .catch((e: Error) => notify.error(e.message));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const card = (label: string, value: number | undefined) => (
    <div className="card">
      <div className="card-label">{label}</div>
      <div className="card-value" data-testid={`stat-${label.toLowerCase().replace(/ /g, "-")}`}>
        {value ?? "–"}
      </div>
    </div>
  );

  return (
    <section>
      <h1>Dashboard</h1>
      <div className="cards">
        {card("Total Users", stats?.total)}
        {card("Active Users", stats?.active)}
        {card("Inactive Users", stats?.inactive)}
      </div>
    </section>
  );
}
