import type { Metadata } from "next";

import { PreviewForm } from "@techshop/ui/components";

export const metadata: Metadata = { title: "Your account | TechShop Staff" };

export default function AccountPage() {
  return (
    <div className="stack stack--lg">
      <div className="stack stack--sm">
        <h1 className="text-xl">Your account</h1>
        <p className="text-sm text-muted">
          Operations Manager (sample role) · Roles and permissions arrive with
          staff sign-in.
        </p>
      </div>
      <div className="card card--roomy">
        <PreviewForm
          id="profile"
          groups={[
            {
              legend: "Profile",
              fields: [
                {
                  name: "name",
                  label: "Full name",
                  required: true,
                  autoComplete: "name",
                  half: true,
                },
                {
                  name: "phone",
                  label: "Phone number",
                  type: "tel",
                  autoComplete: "tel",
                  half: true,
                },
                {
                  name: "email",
                  label: "Work email",
                  type: "email",
                  required: true,
                  autoComplete: "email",
                },
              ],
            },
            {
              legend: "Change password",
              fields: [
                {
                  name: "current",
                  label: "Current password",
                  type: "password",
                  autoComplete: "current-password",
                },
                {
                  name: "password",
                  label: "New password",
                  type: "password",
                  autoComplete: "new-password",
                  minLength: 8,
                  half: true,
                },
                {
                  name: "confirm",
                  label: "Confirm new password",
                  type: "password",
                  autoComplete: "new-password",
                  matches: "password",
                  half: true,
                },
              ],
            },
          ]}
          submitLabel="Save changes"
          successTitle="Changes ready"
          successMessage="In the live portal your profile would now be updated."
        />
      </div>
    </div>
  );
}
