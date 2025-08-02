import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
	// Get token from cookies
	const token = request.cookies.get("jwt_token")?.value;
	const pathname = request.nextUrl.pathname;

	// Public paths that don't require authentication
	const publicPaths = ["/login", "/callback", "/error"];
	const isPublicPath = publicPaths.some((path) => pathname.startsWith(path));

	// If user is authenticated and trying to access login page, redirect to home
	if (token && pathname === "/login") {
		return NextResponse.redirect(new URL("/", request.url));
	}

	// If accessing public path, allow
	if (isPublicPath) {
		return NextResponse.next();
	}

	// If no token and not on public path, redirect to login
	if (!token) {
		const loginUrl = new URL("/login", request.url);
		// Preserve the original URL to redirect back after login
		loginUrl.searchParams.set("redirect", pathname);
		return NextResponse.redirect(loginUrl);
	}

	// If has token, allow access
	return NextResponse.next();
}

export const config = {
	matcher: [
		/*
		 * Match all request paths except for the ones starting with:
		 * - api (API routes)
		 * - _next/static (static files)
		 * - _next/image (image optimization files)
		 * - favicon.ico (favicon file)
		 */
		"/((?!api|_next/static|_next/image|favicon.ico).*)",
	],
};
