import { useState } from "react";
import {
  Megaphone,
  Target,
  DollarSign,
  Shield,
  BarChart3,
  Zap,
  Users,
  MessageSquare,
  TrendingUp,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface OnboardingPageProps {
  onComplete: (mode: "publisher" | "advertiser") => void;
}

const publisherBenefits = [
  {
    icon: DollarSign,
    title: "Monetize Your Audience",
    description: "Turn your channel's reach into revenue",
  },
  {
    icon: Shield,
    title: "Secure Escrow",
    description: "Payments held safely until ads are delivered",
  },
  {
    icon: BarChart3,
    title: "Analytics Dashboard",
    description: "Track performance and optimize earnings",
  },
  {
    icon: Zap,
    title: "Quick Setup",
    description: "Add our bot and start receiving offers",
  },
];

const advertiserBenefits = [
  {
    icon: Users,
    title: "Targeted Reach",
    description: "Connect with audiences that match your goals",
  },
  {
    icon: Shield,
    title: "Protected Payments",
    description: "Pay only when ads are successfully posted",
  },
  {
    icon: MessageSquare,
    title: "Direct Negotiation",
    description: "Chat with publishers to craft perfect ads",
  },
  {
    icon: TrendingUp,
    title: "Campaign Insights",
    description: "Measure ROI with detailed analytics",
  },
];

const isPublisher = (mode: string) => mode === "publisher";

export function OnboardingPage({ onComplete }: OnboardingPageProps) {
  const [mode, setMode] = useState<"publisher" | "advertiser">("publisher");
  const [saving, setSaving] = useState(false);

  const benefits = isPublisher(mode) ? publisherBenefits : advertiserBenefits;

  const handleJoin = async () => {
    setSaving(true);
    try {
      await onComplete(mode);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div
      className="min-h-screen bg-background flex flex-col"
      style={{ paddingTop: "var(--total-safe-area-top, 0px)" }}
    >
      <div className="flex flex-col max-w-md mx-auto w-full p-4">
        <div className="text-center pt-6 pb-8">
          <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-primary/10 flex items-center justify-center">
            <Megaphone className="h-8 w-8 text-primary" />
          </div>
          <h1 className="text-2xl font-bold">Ad Marketplace</h1>
          <p className="text-muted-foreground mt-1">Connect & monetize ads</p>
        </div>

        <div>
          <p className="text-lg italic text-center text-foreground mb-3">
            I am a... 🤔
          </p>
          <div className="relative">
            <div className="flex relative z-10">
              <button
                type="button"
                onClick={() => setMode("publisher")}
                className={cn(
                  "flex flex-1 items-center justify-center gap-2 py-3 text-sm font-medium transition-colors rounded-t-lg",
                  isPublisher(mode)
                    ? "bg-card text-primary border-2 border-primary border-b-card -mb-[2px]"
                    : "bg-muted text-muted-foreground hover:text-foreground",
                )}
              >
                <Megaphone className="h-4 w-4" />
                Publisher
              </button>
              <button
                type="button"
                onClick={() => setMode("advertiser")}
                className={cn(
                  "flex flex-1 items-center justify-center gap-2 py-3 text-sm font-medium transition-colors rounded-t-lg",
                  !isPublisher(mode)
                    ? "bg-card text-primary border-2 border-primary border-b-card -mb-[2px]"
                    : "bg-muted text-muted-foreground hover:text-foreground",
                )}
              >
                <Target className="h-4 w-4" />
                Advertiser
              </button>
            </div>
            <div className="bg-card p-4 space-y-4 border-2 border-primary rounded-b-lg">
              {benefits.map((benefit) => (
                <div key={benefit.title} className="flex gap-4">
                  <div className="flex-shrink-0 w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center">
                    <benefit.icon className="h-5 w-5 text-primary" />
                  </div>
                  <div>
                    <h3 className="font-medium">{benefit.title}</h3>
                    <p className="text-sm text-muted-foreground">
                      {benefit.description}
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        <div className="mt-8 pb-4">
          <Button
            className="w-full h-auto py-4 relative"
            size="lg"
            onClick={handleJoin}
            disabled={saving}
          >
            <div className="flex flex-col gap-1">
              <span className="text-lg font-semibold">
                {saving
                  ? "Joining..."
                  : `Join as ${isPublisher(mode) ? "Publisher" : "Advertiser"}`}
              </span>
              <span className="text-xs text-primary-foreground/60 font-normal">
                (don't worry, you can always change this later)
              </span>
            </div>
            {!saving && (
              <span className="absolute right-4 top-1/2 -translate-y-1/2 text-2xl font-bold">
                →
              </span>
            )}
          </Button>
        </div>
      </div>
    </div>
  );
}
