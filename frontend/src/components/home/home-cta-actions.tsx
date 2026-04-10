"use client";

import Link from "next/link";
import { ArrowRight, MessageCircle } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/contexts/auth";

export function HomeCtaActions() {
	const { isAuthenticated, loading } = useAuth();

	if (loading) {
		return (
			<div className="flex flex-col gap-3 sm:flex-row">
				<Button size="lg" disabled className="min-w-52">
					Đang kiểm tra phiên...
				</Button>
				<Button size="lg" variant="outline" asChild>
					<Link href="/login">Đăng nhập với Google</Link>
				</Button>
			</div>
		);
	}

	if (isAuthenticated) {
		return (
			<div className="flex flex-col gap-3 sm:flex-row">
				<Button size="lg" asChild className="min-w-52">
					<Link href="/">
						Vào phòng chat
						<MessageCircle />
					</Link>
				</Button>
				<Button size="lg" variant="outline" asChild>
					<Link href="#faq">
						Xem FAQ
						<ArrowRight />
					</Link>
				</Button>
			</div>
		);
	}

	return (
		<div className="flex flex-col gap-3 sm:flex-row">
			<Button size="lg" asChild className="min-w-52">
				<Link href="/login">
					Đăng nhập để bắt đầu
					<ArrowRight />
				</Link>
			</Button>
			<Button size="lg" variant="outline" asChild>
				<Link href="#cach-hoat-dong">Xem cách hoạt động</Link>
			</Button>
		</div>
	);
}
