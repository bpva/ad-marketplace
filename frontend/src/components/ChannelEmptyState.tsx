import { Plus, Settings, Users, Bot } from "lucide-react";
import { Button } from "@/components/ui/button";

const steps = [
  {
    icon: Settings,
    text: "Open channel settings",
  },
  {
    icon: Users,
    text: "Go to Administrators",
  },
  {
    icon: Bot,
    text: "Add @adxchange_bot",
    detail: 'with "Post messages" permission',
  },
];

export function ChannelEmptyState() {
  return (
    <div className="flex-1 flex flex-col items-center justify-center p-4">
      <div className="max-w-sm w-full text-center">
        <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-primary/10 flex items-center justify-center">
          <Plus className="h-8 w-8 text-primary" />
        </div>

        <h1 className="text-xl font-bold mb-2">Add Your First Channel</h1>
        <p className="text-muted-foreground text-sm mb-6">
          Connect your Telegram channel to start receiving ad offers
        </p>

        <div className="bg-card rounded-xl border border-border p-4 text-left space-y-4 mb-6">
          {steps.map((step, i) => (
            <div key={i} className="flex gap-3 items-start">
              <div className="flex-shrink-0 w-8 h-8 rounded-lg bg-primary/10 flex items-center justify-center">
                <step.icon className="h-4 w-4 text-primary" />
              </div>
              <div className="pt-1">
                <p className="text-sm font-medium">{step.text}</p>
                {step.detail && (
                  <p className="text-xs text-muted-foreground mt-0.5">{step.detail}</p>
                )}
              </div>
            </div>
          ))}
        </div>

        <Button asChild className="w-full">
          <a href="https://t.me/adxchange_bot">Open @adxchange_bot</a>
        </Button>

        <p className="text-xs text-muted-foreground mt-4">
          Channels appear for the owner first. As owner, you can invite others to manage.
        </p>
      </div>
    </div>
  );
}
