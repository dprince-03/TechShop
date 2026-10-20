import Link from "next/link";

import { EmptyState } from "@techshop/ui/components";

/** Shown in place of live data when the API can't be reached. Never a dead end. */
export function ServiceUnavailable({
  what = "products",
  retryHref,
}: {
  what?: string;
  retryHref: string;
}) {
  return (
    <main id="main" className="section section--page">
      <div className="container container--narrow">
        <EmptyState
          icon="support"
          title={`We can’t load ${what} right now`}
          action={
            <Link href={retryHref} className="btn btn--secondary">
              Try again
            </Link>
          }
        >
          Our shop service isn’t responding. Please try again in a moment.
        </EmptyState>
      </div>
    </main>
  );
}
