"use client";

import { useState } from "react";
import { Sheet, SheetContent, SheetHeader, SheetTitle } from "@/components/ui/sheet";
import { BannedWordsTab } from "./banned-words-tab";
import { ReportsTab } from "./reports-tab";

interface AdminPanelProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export function AdminPanel({ open, onOpenChange }: AdminPanelProps) {
	const [activeTab, setActiveTab] = useState<"words" | "reports">("words");

	return (
		<Sheet open={open} onOpenChange={onOpenChange}>
			<SheetContent side="right" className="w-full sm:max-w-lg flex flex-col overflow-hidden p-0">
				<div className="px-6 pt-6 pb-0 border-b">
					<SheetHeader className="mb-4">
						<SheetTitle>Admin Panel</SheetTitle>
					</SheetHeader>
					<div className="flex">
						<button
							className={`px-4 py-2 text-sm font-medium transition-colors ${
								activeTab === "words"
									? "border-b-2 border-primary text-foreground"
									: "text-muted-foreground hover:text-foreground"
							}`}
							onClick={() => setActiveTab("words")}
						>
							Banned Words
						</button>
						<button
							className={`px-4 py-2 text-sm font-medium transition-colors ${
								activeTab === "reports"
									? "border-b-2 border-primary text-foreground"
									: "text-muted-foreground hover:text-foreground"
							}`}
							onClick={() => setActiveTab("reports")}
						>
							Reports
						</button>
					</div>
				</div>
				<div className="flex-1 overflow-y-auto px-6 py-4">
					{activeTab === "words" ? <BannedWordsTab /> : <ReportsTab />}
				</div>
			</SheetContent>
		</Sheet>
	);
}
