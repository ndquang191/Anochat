import type React from "react";
import { cookies } from "next/headers";
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import { AppShellSidebar } from "@/components/app-shell-sidebar";
import AppHeader from "@/components/app-header";
import { AdminProvider } from "@/contexts/admin";

export default async function MainLayout({ children }: { children: React.ReactNode }) {
	const cookieStore = await cookies();
	const sidebarState = cookieStore.get("sidebar_state");
	const defaultOpen = !sidebarState || sidebarState.value === "true";

	return (
		<SidebarProvider defaultOpen={defaultOpen}>
			<AdminProvider>
				<AppShellSidebar />
				<SidebarInset className="h-screen flex flex-col overflow-hidden">
					<AppHeader trigger={<SidebarTrigger className="-ml-1" />} />
					<main className="flex-1 mt-16 overflow-hidden">{children}</main>
				</SidebarInset>
			</AdminProvider>
		</SidebarProvider>
	);
}
