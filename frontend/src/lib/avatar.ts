export function getAvatarUrl(telegramId: number, photoUrl?: string): string {
  if (photoUrl) {
    return photoUrl;
  }
  return `https://api.dicebear.com/9.x/thumbs/svg?seed=${telegramId}`;
}
