import { useRef, useState } from "react";
import { createPortal } from "react-dom";
import type { Player, PlayerNewsUpdate } from "../shared/api/types";
import { Button } from "../shared/ui/Button";
import { Dialog } from "../shared/ui/Dialog";

interface PlayerNewsDetailsProps {
  player: Player;
  update: PlayerNewsUpdate;
}

export function PlayerNewsDetails({ player, update }: PlayerNewsDetailsProps) {
  const [open, setOpen] = useState(false);
  const closeButton = useRef<HTMLButtonElement>(null);
  const status = update.availability;
  const dialogTitle = `player-news-${player.id}`;
  const statusText =
    status && !isActive(status.status) ? availabilityLabel(status) : update.events.length ? "Latest news" : "";

  if (!statusText) return null;

  return (
    <>
      <span className="player-news-summary">
        <span className={status && !isActive(status.status) ? "availability-alert" : "news-available"}>
          {statusText}
        </span>
        <Button
          variant="neutral"
          className="inline-details-button"
          aria-label={`View news details for ${player.name}`}
          onClick={() => setOpen(true)}
        >
          Details
        </Button>
      </span>
      {createPortal(
        <Dialog open={open} labelledBy={dialogTitle} initialFocusRef={closeButton} onClose={() => setOpen(false)}>
          <div className="dialog-heading">
            <div>
              <p className="eyebrow">Player availability and news</p>
              <h2 id={dialogTitle}>{player.name}</h2>
            </div>
            <Button ref={closeButton} onClick={() => setOpen(false)}>
              Close
            </Button>
          </div>
          {status ? (
            <section aria-labelledby={`${dialogTitle}-status`}>
              <h3 id={`${dialogTitle}-status`}>Current availability</h3>
              <dl className="news-status-details">
                <div>
                  <dt>Status</dt>
                  <dd>{status.status}</dd>
                </div>
                {status.injury ? (
                  <div>
                    <dt>Injury</dt>
                    <dd>{status.injury}</dd>
                  </div>
                ) : null}
                {status.practiceParticipation ? (
                  <div>
                    <dt>Practice</dt>
                    <dd>{status.practiceParticipation}</dd>
                  </div>
                ) : null}
                <div>
                  <dt>Source</dt>
                  <dd>{status.sourceName}</dd>
                </div>
                <div>
                  <dt>Updated</dt>
                  <dd>{formatNewsTime(status.updatedAt)}</dd>
                </div>
              </dl>
            </section>
          ) : null}
          <section aria-labelledby={`${dialogTitle}-stories`}>
            <h3 id={`${dialogTitle}-stories`}>Recent updates</h3>
            {update.events.length ? (
              <ul className="player-news-list">
                {update.events.map((event) => (
                  <li key={event.id}>
                    {event.url ? (
                      <a href={event.url} target="_blank" rel="noreferrer">
                        {event.title}
                      </a>
                    ) : (
                      <span>{event.title}</span>
                    )}
                    <span>
                      {event.sourceName} · {formatNewsTime(event.publishedAt)} · {event.confidence} confidence
                    </span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="empty-state">No recent player-specific stories.</p>
            )}
          </section>
        </Dialog>,
        document.body,
      )}
    </>
  );
}

function availabilityLabel(status: NonNullable<PlayerNewsUpdate["availability"]>): string {
  return status.injury ? `${status.status} · ${status.injury}` : status.status;
}

function isActive(status: string): boolean {
  return status.trim().toLowerCase() === "active";
}

function formatNewsTime(value: string): string {
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
}
