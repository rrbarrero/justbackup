import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { jwtVerify } from "jose";

const publicPaths = ["/login", "/setup"];

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // 1. Allow static assets and internal next paths immediately
  if (
    pathname.startsWith("/_next") ||
    pathname.startsWith("/static") ||
    pathname.includes(".")
  ) {
    return NextResponse.next();
  }

  // 2. Check Setup Status (only for non-API routes)
  const isApi = pathname.startsWith("/api");
  const API_URL = process.env.BACKEND_INTERNAL_URL;

  if (!isApi && API_URL) {
    try {
      const setupRes = await fetch(`${API_URL}/api/v1/setup-status`);
      const { setupRequired } = await setupRes.json();

      if (setupRequired && pathname !== "/setup") {
        return NextResponse.redirect(new URL("/setup", request.url));
      }
      if (!setupRequired && pathname === "/setup") {
        return NextResponse.redirect(new URL("/login", request.url));
      }
    } catch (e) {
      console.error("Proxy: Failed to check setup status", e);
    }
  }

  // 3. Allow public paths and API routes (token addition happens later)
  if (isApi || publicPaths.some((path) => pathname.startsWith(path))) {
    // If it's an API route, we still want to add the token if available,
    // but we don't want to redirect if it's missing (let the backend return 401)
    if (isApi) {
      const token = request.cookies.get("token")?.value;
      if (token) {
        try {
          const jwtSecret = process.env.JWT_SECRET;
          if (jwtSecret) {
            const secret = new TextEncoder().encode(jwtSecret);
            await jwtVerify(token, secret);
            const requestHeaders = new Headers(request.headers);
            requestHeaders.set("Authorization", `Bearer ${token}`);
            return NextResponse.next({
              request: { headers: requestHeaders },
            });
          }
        } catch (err) {
          // Token invalid, let it pass without header (backend will handle)
        }
      }
      return NextResponse.next();
    }
    return NextResponse.next();
  }

  // 4. Token Validation for protected routes
  const token = request.cookies.get("token")?.value;

  if (!token) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  try {
    const jwtSecret = process.env.JWT_SECRET;
    if (!jwtSecret) {
      return NextResponse.redirect(new URL("/login", request.url));
    }
    const secret = new TextEncoder().encode(jwtSecret);

    await jwtVerify(token, secret);

    const requestHeaders = new Headers(request.headers);
    requestHeaders.set("Authorization", `Bearer ${token}`);

    return NextResponse.next({
      request: {
        headers: requestHeaders,
      },
    });
  } catch (err) {
    return NextResponse.redirect(new URL("/login", request.url));
  }
}

export const config = {
  matcher: [
    /*
     * Match all request paths except for the ones starting with:
     * - api (API routes) -> actually we might want to protect API routes too, but let's start with pages
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     */
    "/((?!_next/static|_next/image|favicon.ico).*)",
  ],
};
