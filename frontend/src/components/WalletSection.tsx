import { useEffect, useRef } from "react";
import { Wallet, ExternalLink, X } from "lucide-react";
import { useTonConnectUI, useTonWallet } from "@tonconnect/ui-react";
import { Button } from "@/components/ui/button";
import { linkWallet, unlinkWallet } from "@/lib/api";
import { truncateAddress } from "@/lib/format";

export function WalletSection() {
  const [tonConnectUI] = useTonConnectUI();
  const wallet = useTonWallet();
  const prevAddress = useRef<string | undefined>(undefined);

  useEffect(() => {
    const address = wallet?.account?.address;
    if (address === prevAddress.current) return;
    prevAddress.current = address;

    if (address) {
      linkWallet(address).catch(() => {});
    }
  }, [wallet?.account?.address]);

  const handleConnect = () => {
    tonConnectUI.openModal();
  };

  const handleDisconnect = async () => {
    await tonConnectUI.disconnect();
    unlinkWallet().catch(() => {});
  };

  const address = wallet?.account?.address;
  const explorerUrl = address ? `https://tonviewer.com/${address}` : undefined;

  return (
    <div className="bg-card rounded-lg border border-border p-4 space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">Wallet</h2>
        {wallet && (
          <span className="flex items-center gap-1.5 text-xs font-medium text-green-600 dark:text-green-500">
            <span className="h-2 w-2 rounded-full bg-green-500" />
            Connected
          </span>
        )}
      </div>

      {!wallet ? (
        <>
          <Button onClick={handleConnect} className="w-full gap-2">
            <Wallet className="h-4 w-4" />
            Connect Wallet
          </Button>
          <p className="text-xs text-muted-foreground">
            Connect your TON wallet to pay for ads or receive revenue
          </p>
        </>
      ) : (
        <div className="flex items-center justify-between">
          <a
            href={explorerUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-2 text-sm font-mono text-primary hover:underline"
          >
            {truncateAddress(address ?? "")}
            <ExternalLink className="h-3.5 w-3.5" />
          </a>
          <Button variant="ghost" size="icon" onClick={handleDisconnect}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      )}
    </div>
  );
}
