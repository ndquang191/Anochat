import type React from "react";
import { cookies } from "next/headers";
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import "../globals.css"; // Ensure your global CSS is imported
import Header from "@/components/header";
import { ConnectionProvider } from "@/hooks/use-connection";
import { AlertDialogProvider } from "@/components/alert-dialog-provider";
import { Toaster } from "sonner";

export default async function RootLayout({ children }: { children: React.ReactNode }) {
	const cookieStore = await cookies();
	const sidebarState = cookieStore.get("sidebar_state");
	// Set defaultOpen to true if no cookie exists or if the cookie value is "true"
	const defaultOpen = !sidebarState || sidebarState.value === "true";

	return (
		<html lang="en">
			<body>
				<AlertDialogProvider>
					<ConnectionProvider>
						<SidebarProvider defaultOpen={defaultOpen}>
							<AppSidebar />
							<SidebarInset className="h-screen flex flex-col overflow-hidden">
								<Header trigger={<SidebarTrigger className="-ml-1" />} />
								<main className="flex-1 mt-16 overflow-hidden">{children}</main>
							</SidebarInset>
						</SidebarProvider>
					</ConnectionProvider>
				</AlertDialogProvider>
				<Toaster position="top-right" />
			</body>
		</html>
	);
}
