import Link from "next/link";
import { AuthForm } from "@/components/AuthForm";

export default function LoginPage() {
  return (
    <main className="mx-auto w-full max-w-xl px-4 py-10 md:px-8">
      <section className="rounded-lg border border-[rgba(77,100,124,0.16)] bg-white p-6 shadow-soft">
        <h1 className="text-3xl font-semibold text-marine-text">登录</h1>
        <AuthForm mode="login" />
        <p className="mt-5 text-center text-sm text-marine-muted">
          还没有账号？{" "}
          <Link className="font-semibold text-marine-deep" href="/register">
            去注册
          </Link>
        </p>
      </section>
    </main>
  );
}
