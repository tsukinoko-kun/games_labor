import { ZodIssue } from "zod";
import type { PlayerData } from "./types";

export const seed = Math.random();

const names = [
  "Alex",
  "Avery",
  "Bailey",
  "Blake",
  "Charlie",
  "Dakota",
  "Drew",
  "Ellis",
  "Emerson",
  "Finley",
  "Harper",
  "Jamie",
  "Jordan",
  "Kai",
  "Kennedy",
  "Morgan",
  "Quinn",
  "Riley",
  "Rowan",
  "Sawyer",
];

const appearances = [
  "Mittlere Statur, unauffällige Kleidung",
  "Schlanke Silhouette, praktische Adjustierung",
  "Kräftiger Körperbau, einfache Gewandung",
  "Androgyner Look, neutrale Farben",
  "Gewöhnliches Aussehen, strapazierfähige Stoffe",
];

const origins = [
  "Aus den Gassen einer geschäftigen Metropole, Techniker",
  "Vom mystischen Orden der Sternen-Nomaden, Wissenshüter",
  "Aus einer abgeschiedenen Agrarkolonie, Überlebenskünstler",
  "Von der angesehenen Handelsgilde, Händler",
  "Aus den Reihen einer Verteidigungsmiliz, Soldat",
];

export function seededRandomName() {
  return names[Math.floor(seed * 4321) % names.length];
}

export function seededRandomAppearance() {
  return appearances[Math.floor(seed * 13) % appearances.length];
}

export function seededRandomOrigin() {
  return origins[Math.floor(seed * 1234) % origins.length];
}

export function seededRandomFloat(min: number, max: number) {
  return ((seed * 589123) % (max - min + 1)) + min;
}

export function seededRandomInt(min: number, max: number) {
  return Math.floor(seededRandomFloat(min, max));
}

export function seededRandomCharacter() {
  return {
    name: seededRandomName(),
    age: seededRandomInt(16, 52) + " Jahre",
    origin: seededRandomOrigin(),
    appearance: seededRandomAppearance(),
  } satisfies PlayerData;
}

export function getCookie(name: string) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return (parts.pop() ?? "").split(";").shift();
  return null;
}

export const myUserId = getCookie("user_id") ?? "";

export function stringToColor(str: string) {
  let hash = 0;

  // Generate hash
  for (let i = 0; i < str.length; i++) {
    hash = (hash << 5) - hash + str.charCodeAt(i);
    hash = hash & hash; // Convert to 32bit integer
  }

  // Create separate hash values for R, G, B
  const r = (hash & 0xff0000) >> 16;
  const g = (hash & 0x00ff00) >> 8;
  const b = hash & 0x0000ff;

  // Scale to mid-range (60-200) to avoid too dark or too bright
  const scaleToMidRange = (value: number) => {
    return Math.floor(60 + (value % 141));
  };

  // Format as hex
  return `#${scaleToMidRange(r).toString(16).padStart(2, "0")}${scaleToMidRange(
    g,
  )
    .toString(16)
    .padStart(2, "0")}${scaleToMidRange(b).toString(16).padStart(2, "0")}`;
}

export function zodErr(err: ZodIssue): string {
  switch (err.code) {
    case "invalid_type":
      return `invalid type ${err.received} for ${err.path.join(".")}, expected ${err.expected}.`;
    case "invalid_literal":
      return `invalid literal ${err.received} for ${err.path.join(".")}, expected ${err.expected}.`;
    case "custom":
      return `custom error ${err.message} at ${err.path.join(".")}.`;
    case "invalid_union": {
      const mainMsg = `invalid union at ${err.path.join(".")}`;
      if (err.unionErrors && err.unionErrors.length > 0) {
        return (
          mainMsg +
          "\n" +
          err.unionErrors
            .map((err) => err.issues.map(zodErr).join(". ") + ".")
            .join("\n")
        );
      }
      return mainMsg;
    }
    case "invalid_union_discriminator":
      return `invalid union at ${err.path.join(".")}.`;
    case "invalid_enum_value":
      return `invalid enum value ${err.received} for ${err.path.join(".")}.`;
    case "unrecognized_keys":
      return `unrecognized keys ${err.keys.join(", ")} for ${err.path.join(".")}.`;
    case "invalid_arguments": {
      const mainMsg = `invalid arguments ${err.path.join(".")}`;
      if (
        err.argumentsError &&
        err.argumentsError.errors &&
        err.argumentsError.errors.length > 0
      ) {
        return (
          mainMsg + "\n" + err.argumentsError.errors.map(zodErr).join("\n")
        );
      }
      return mainMsg + ".";
    }
    case "invalid_return_type": {
      const mainMsg = `invalid return type for ${err.path.join(".")}`;
      if (
        err.returnTypeError &&
        err.returnTypeError.errors &&
        err.returnTypeError.errors.length > 0
      ) {
        return (
          mainMsg + "\n" + err.returnTypeError.errors.map(zodErr).join("\n")
        );
      }
      return mainMsg + ".";
    }
    case "invalid_date":
      return `invalid date at ${err.path.join(".")}.`;
    case "invalid_string":
      return `invalid string at ${err.path.join(".")}.`;
    case "too_small": {
      let msg = `value at ${err.path.join(".")} bust be `;
      if (err.inclusive) msg += ` >= ${err.minimum}`;
      else if (err.exact) msg += ` = ${err.minimum}`;
      else msg += ` > ${err.minimum}`;
      msg += ".";
      return msg;
    }
    case "too_big": {
      let msg = `value at ${err.path.join(".")} bust be `;
      if (err.inclusive) msg += ` <= ${err.maximum}`;
      else if (err.exact) msg += ` = ${err.maximum}`;
      else msg += ` < ${err.maximum}`;
      msg += ".";
      return msg;
    }
    case "invalid_intersection_types":
      `invalid intersection types at ${err.path.join(".")}.`;
    case "not_multiple_of":
      return `value at ${err.path.join(".")} is not multiple.`;
    case "not_finite":
      return `value at ${err.path.join(".")} is not finite.`;
  }
}
