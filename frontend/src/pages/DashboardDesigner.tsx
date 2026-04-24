/**
 * Stub — the real designer lands in Phase N. Keeping this module so
 * App.tsx can lazy-import the `/app/dashboards/:id/edit` target today.
 */
export default function DashboardDesigner() {
  return (
    <div className="flex min-h-[40vh] items-center justify-center p-6 text-center">
      <div className="max-w-md space-y-2">
        <h2 className="text-lg font-semibold">Designer coming soon</h2>
        <p className="text-sm text-muted-foreground">
          The dashboard designer is under construction. It will let you add,
          move, resize and configure widgets directly on this page.
        </p>
      </div>
    </div>
  );
}
