import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import Link from "next/link";

export default async function Page({ searchParams }: { searchParams: { error?: string } }) {
	console.error("[Auth Error]", searchParams?.error || "Unspecified error");

	return (
		<div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
			<div className="w-full max-w-sm">
				<div className="flex flex-col gap-6">
					<Card>
						<CardHeader>
							<CardTitle className="text-2xl">Authentication Error</CardTitle>
						</CardHeader>
						<CardContent className="flex flex-col gap-4">
							{searchParams?.error ? (
								<p className="text-sm text-destructive-500 bg-destructive-50 p-3 rounded">
									{searchParams.error}
								</p>
							) : (
								<p className="text-sm text-muted-foreground">
									An unexpected error occurred during authentication.
								</p>
							)}

							<div className="flex flex-col gap-2">
								<Button asChild>
									<Link href="/login">Try Again</Link>
								</Button>
								<Button variant="outline" asChild>
									<Link href="/">Back to Home</Link>
								</Button>
							</div>

							<p className="text-xs text-muted-foreground mt-4">
								If this error persists, please clear your browser cookies and try again. If
								the problem continues, contact support.
							</p>
						</CardContent>
					</Card>
				</div>
			</div>
		</div>
	);
}
