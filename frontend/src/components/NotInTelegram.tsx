import { MessageCircle } from "lucide-react";
import { Button } from "@/components/ui/button";

export function NotInTelegram() {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center p-6 text-center bg-background text-foreground">
      <div className="mb-6 p-4 rounded-full bg-primary/10">
        <MessageCircle className="w-12 h-12 text-primary" />
      </div>
      <h1 className="text-2xl font-bold mb-3">Open in Telegram</h1>
      <p className="text-muted-foreground max-w-sm mb-6">
        This app is designed to work as a Telegram Mini App. Please open it
        through Telegram to continue.
      </p>
      <Button asChild>
        <a href="https://t.me/adxchange_bot">Open Telegram</a>
      </Button>
    </div>
  );
}
