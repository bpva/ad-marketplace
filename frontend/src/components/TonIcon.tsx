interface TonIconProps {
  size?: number;
  className?: string;
}

export function TonIcon({ size = 16, className }: TonIconProps) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      width={size}
      height={size}
      fill="currentColor"
      className={className}
    >
      <path
        d="M12 0C5.373 0 0 5.373 0 12s5.373 12 12 12 12-5.373 12-12S18.627 0 12 0z"
        opacity=".15"
      />
      <path d="M7.902 6.697h8.196c1.505 0 2.462 1.61 1.727 2.905l-5.078 8.93a.924.924 0 0 1-1.611-.007l-4.96-8.923c-.724-1.3.23-2.905 1.726-2.905zm3.212 2.069-3.036 5.29a.462.462 0 0 0 .4.693h2.636V8.766zm1.772 0v5.983h2.636a.462.462 0 0 0 .4-.692l-3.036-5.291z" />
    </svg>
  );
}
