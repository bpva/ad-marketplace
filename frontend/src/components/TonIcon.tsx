interface TonIconProps {
  className?: string;
}

export function TonIcon({ className }: TonIconProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="12 15 32 30"
      fill="currentColor"
      className={className}
    >
      <path d="M37.6,15.6H18.4c-3.5,0-5.7,3.8-4,6.9l11.8,20.5c0.8,1.3,2.7,1.3,3.5,0l11.8-20.5C43.3,19.4,41.1,15.6,37.6,15.6z M26.3,36.8l-2.6-5l-6.2-11.1c-0.4-0.7,0.1-1.6,1-1.6h7.8V36.8z M38.5,20.7l-6.2,11.1l-2.6,5V19.1h7.8C38.4,19.1,38.9,20,38.5,20.7z" />
    </svg>
  );
}
