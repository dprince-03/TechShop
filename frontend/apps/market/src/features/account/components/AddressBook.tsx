"use client";

import { useState, useTransition } from "react";

import type { Schema } from "@techshop/api-client";

import { deleteAddress, type FormState } from "../actions";

import styles from "./Account.module.css";
import { AddressForm } from "./AddressForm";
import { FormFeedback } from "./FormFeedback";

type Address = Schema<"IdentityAddress">;

export function formatAddress(a: Address, stateName?: string) {
  return [a.line1, a.line2, a.city, a.lga, stateName ?? a.stateCode]
    .filter(Boolean)
    .join(", ");
}

export function AddressBook({
  addresses,
  states,
}: {
  addresses: Address[];
  states: Schema<"IdentityNigerianState">[];
}) {
  const [editing, setEditing] = useState<string | "new" | null>(
    addresses.length ? null : "new",
  );
  const [status, setStatus] = useState<FormState>({});
  const [pending, startTransition] = useTransition();
  const stateName = (code: string) => states.find((s) => s.code === code)?.name;

  return (
    <div className="stack">
      <FormFeedback state={status} />
      {addresses.length ? (
        <ul role="list" className={styles.list}>
          {addresses.map((a) => (
            <li key={a.id} className={styles.item}>
              {editing === a.id ? (
                <AddressForm
                  states={states}
                  address={a}
                  onDone={() => setEditing(null)}
                />
              ) : (
                <>
                  <div className="stack stack--sm">
                    <p className={styles.itemTitle}>
                      {a.recipientName}
                      {a.isDefault ? (
                        <span className="badge badge--outline">Default</span>
                      ) : null}
                    </p>
                    <p className="text-sm text-muted">
                      {formatAddress(a, stateName(a.stateCode))}
                    </p>
                    {a.landmark ? (
                      <p className="text-xs text-muted">
                        Landmark: {a.landmark}
                      </p>
                    ) : null}
                  </div>
                  <div className="cluster">
                    <button
                      type="button"
                      className="btn btn--secondary btn--sm"
                      onClick={() => setEditing(a.id)}
                    >
                      Edit
                    </button>
                    <button
                      type="button"
                      className="btn btn--ghost btn--sm"
                      disabled={pending}
                      onClick={() => {
                        if (
                          !window.confirm(
                            `Remove the address for ${a.recipientName}?`,
                          )
                        )
                          return;
                        startTransition(async () =>
                          setStatus(await deleteAddress(a.id)),
                        );
                      }}
                    >
                      Remove
                    </button>
                  </div>
                </>
              )}
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-sm text-muted">You haven’t saved an address yet.</p>
      )}
      {editing === "new" ? (
        <div className="card card--flat">
          <AddressForm states={states} onDone={() => setEditing(null)} />
        </div>
      ) : (
        <button
          type="button"
          className="btn btn--secondary"
          onClick={() => setEditing("new")}
        >
          Add an address
        </button>
      )}
    </div>
  );
}
