export interface Health {
  status: "ok";
  service: string;
  version: string;
}

export async function getHealth(): Promise<Health> {
  const response = await fetch("/api/v1/health");
  if (!response.ok) {
    throw new Error(`Health request failed with ${response.status}`);
  }
  return response.json() as Promise<Health>;
}
