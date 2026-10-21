"use client";

import { ActionButton } from "@/components/ActionButton";

import { decideReview } from "../actions";

import styles from "./ReviewActions.module.css";

/** A review's text, wrapped at a readable width inside the table. */
export function ReviewText({ children }: { children: string }) {
  return <span className={`${styles.text} line-clamp`}>{children}</span>;
}

export function ReviewActions({ id }: { id: string }) {
  return (
    <div className="cluster">
      <ActionButton
        label="Publish"
        variant="inverse"
        action={() => decideReview(id, "publish")}
      />
      <ActionButton label="Hide" action={() => decideReview(id, "hide")} />
      <ActionButton
        label="Reject"
        variant="ghost"
        confirm="Reject this review? It won’t be shown."
        action={() => decideReview(id, "reject")}
      />
    </div>
  );
}
