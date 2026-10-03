import Link from "next/link";

import { EmptyState } from "@techshop/ui/components";

export default function NotFound() {
  return (
    <div className="section--page">
      <div className="container container--narrow">
        <EmptyState
          icon="search"
          title="We couldn’t find that page"
          action={
            <Link href="/" className="btn btn--primary">
              Go to the homepage
            </Link>
          }
        >
          The link may be old, or this page may not exist yet.
        </EmptyState>
      </div>
    </div>
  );
}
