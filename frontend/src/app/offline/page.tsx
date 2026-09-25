import Link from "next/link";
import { WifiOff } from "lucide-react";
import { Button } from "@/components/ui/button";
import { BrandLogo } from "@/components/brand-logo";

export default function OfflinePage() {
	return (
		<main className="flex min-h-svh items-center justify-center bg-background p-6 text-foreground">
			<div className="flex max-w-sm flex-col items-center gap-4 text-center">
				<BrandLogo />
				<WifiOff className="size-8 text-muted-foreground" aria-hidden="true" />
				<h1 className="text-xl font-semibold">Bạn đang ngoại tuyến / You’re offline</h1>
				<p className="text-sm text-muted-foreground">AnoChat cần kết nối mạng để tìm người và gửi tin nhắn.<br />AnoChat needs a connection to match and send messages.</p>
				<Button asChild><Link href="/">Thử lại / Retry</Link></Button>
			</div>
		</main>
	);
}
