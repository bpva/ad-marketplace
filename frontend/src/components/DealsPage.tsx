import { Handshake } from "lucide-react";

export function DealsPage() {
  return (
    <div className="flex-1 flex flex-col items-center justify-center p-4">
      <div className="text-center">
        <div className="w-16 h-16 rounded-2xl bg-muted flex items-center justify-center mx-auto mb-4">
          <Handshake className="h-8 w-8 text-muted-foreground" />
        </div>
        <h1 className="text-xl font-semibold mb-2">Deals</h1>
        <p className="text-muted-foreground text-sm">Coming soon</p>
      </div>
    </div>
  );
}
