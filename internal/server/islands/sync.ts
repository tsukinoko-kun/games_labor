type VersionedData<T> = { version: number; data: T };

export class Sync<T extends object> {
  private data: T;
  private version = 0;
  private versioned: VersionedData<T>;
  private readonly subscribers = new Set<() => void>();

  public constructor(v: T) {
    this.data = v;
    this.versioned = { version: this.version, data: this.data };
  }

  public override(value: T) {
    this.data = value;
    this.notify();
  }

  public set(path: string, value: any) {
    let v: any = this.data;
    const segments = path.split(".");
    if (segments.length === 0) {
      throw new Error(`invalid empty path into data`);
    }
    for (let i = 0; i < segments.length - 1; i++) {
      const segment = segments[i];
      if (typeof v !== "object") {
        throw new Error(
          `invalid path ${path} into data at segment ${segment}: not an object`,
        );
      }
      if (!(segment in v)) {
        throw new Error(
          `invalid path ${path} into data at segment ${segment}: key not found`,
        );
      }
      v = v[segment];
    }
    const segment = segments[segments.length - 1];
    if (!(segment in v)) {
      throw new Error(`invalid path ${path} into data at segment ${segment}`);
    }
    v[segment] = value;
    this.notify();
  }

  public push(path: string, value: any) {
    let v: any = this.data;
    const segments = path.split(".");
    if (segments.length === 0) {
      throw new Error(`invalid empty path into data`);
    }
    for (let i = 0; i < segments.length; i++) {
      const segment = segments[i];
      if (typeof v !== "object") {
        throw new Error(
          `invalid path ${path} into data at segment ${segment}: not an object`,
        );
      }
      if (!(segment in v)) {
        throw new Error(
          `invalid path ${path} into data at segment ${segment}: key not found`,
        );
      }
      v = v[segment];
    }
    if (Array.isArray(v)) {
      v.push(value);
      this.notify();
    } else {
      throw new Error(`element at path ${path} into data is not an array`);
    }
  }

  public subscribe(listender: () => void) {
    this.subscribers.add(listender);
  }

  public unsubscribe(listender: () => void) {
    this.subscribers.delete(listender);
  }

  public getSnapshot() {
    return this.versioned;
  }

  public notify() {
    this.version++;
    this.versioned = { version: this.version, data: this.data };
    (globalThis as any).gameData = this.data;
    for (const sub of this.subscribers) {
      sub();
    }
  }
}
