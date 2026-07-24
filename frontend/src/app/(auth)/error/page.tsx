import { AuthErrorContent } from "@/components/auth-error-content";

export default async function Page({ searchParams }: { searchParams: Promise<{ error?: string }> }) {
	const params = await searchParams;
	console.error("[Auth Error]", params?.error || "Unspecified error");

	return <AuthErrorContent />;
}
