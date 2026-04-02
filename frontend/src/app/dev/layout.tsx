"use client";

import React, { useState, useEffect } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { SidebarProvider, SidebarInset, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import Header from "@/components/header";
import { DevAuthProvider, DevState, DEV_USER, DEV_ROOM, DEV_MESSAGES } from "@/contexts/dev-auth";
import { ThemeProvider } from "@/contexts/theme";
import { AlertDialogProvider } from "@/contexts/alert-dialog";
import { USER_STATE_KEY } from "@/hooks/queries/use-user-state";
import type { UserStateResponse } from "@/types";

// Isolated QueryClient for dev mode — never hits real API
const devQueryClient = new QueryClient({
	defaultOptions: { queries: { retry: false, staleTime: Infinity } },
});

function seedQueryCache(devState: DevState) {
	const data: UserStateResponse = {
		profile: DEV_USER.profile,
		room: devState === "chat" ? DEV_ROOM : undefined,
		messages: devState === "chat" ? DEV_MESSAGES : [],
		in_queue: devState === "queue",
	};
	devQueryClient.setQueryData(USER_STATE_KEY, data);
}

export default function DevLayout({ children }: { children: React.ReactNode }) {
	const [devState, setDevState] = useState<DevState>("lobby");

	// Keep query cache in sync with dev state
	useEffect(() => {
		seedQueryCache(devState);
	}, [devState]);

	// Seed on first render
	seedQueryCache(devState);

	return (
		<ThemeProvider>
			<QueryClientProvider client={devQueryClient}>
				<DevAuthProvider devState={devState}>
					<AlertDialogProvider>
						<SidebarProvider defaultOpen={true}>
							<AppSidebar />
							<SidebarInset className="h-screen flex flex-col overflow-hidden">
								<Header trigger={<SidebarTrigger className="-ml-1" />} />
								<main className="flex-1 mt-16 overflow-hidden">{children}</main>
							</SidebarInset>
						</SidebarProvider>
						{/* Dev state switcher — top right, h-fit */}
						<div className="fixed top-2 right-2 z-50 h-fit flex items-center gap-0.5 bg-black/70 text-white text-[10px] leading-none rounded-md px-1.5 py-1 backdrop-blur-sm shadow-lg">
							<span className="opacity-40 font-mono pr-1.5 border-r border-white/20 mr-0.5">DEV</span>
							{(["lobby", "queue", "chat"] as DevState[]).map((s) => (
								<button
									key={s}
									onClick={() => setDevState(s)}
									className={`px-1.5 py-0.5 rounded leading-none transition-all ${
										devState === s
											? "bg-white/90 text-black font-semibold"
											: "opacity-50 hover:opacity-80"
									}`}
								>
									{s}
								</button>
							))}
						</div>
					</AlertDialogProvider>
				</DevAuthProvider>
			</QueryClientProvider>
		</ThemeProvider>
	);
}
