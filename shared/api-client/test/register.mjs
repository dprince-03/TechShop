// Lets `node --test` run the package's TypeScript sources, which use extensionless imports
// (the style Next.js and Metro expect). Node strips the types itself; this only resolves ".ts".
import { registerHooks } from "node:module";

registerHooks({
  resolve(specifier, context, next) {
    try {
      return next(specifier, context);
    } catch (err) {
      if (err?.code !== "ERR_MODULE_NOT_FOUND" || !/^\.{1,2}\//.test(specifier)) throw err;
      for (const suffix of [".ts", "/index.ts"]) {
        try {
          return next(specifier + suffix, context);
        } catch {
          // try the next candidate
        }
      }
      throw err;
    }
  },
});
