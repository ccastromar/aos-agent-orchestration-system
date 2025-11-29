package main

import (
    "context"
    "errors"
    "testing"
)

type fakeRunner struct{
    ran bool
    err error
}

func (f *fakeRunner) Run(ctx context.Context) error {
    f.ran = true
    return f.err
}

func TestRun_Success(t *testing.T) {
    // arrange
    fr := &fakeRunner{}
    oldCtor := appCtor
    oldFatalf := fatalf
    t.Cleanup(func(){ appCtor = oldCtor; fatalf = oldFatalf })
    appCtor = func() (runner, error) { return fr, nil }

    calledFatal := false
    fatalf = func(format string, v ...any) { calledFatal = true }

    // act
    run(context.Background())

    // assert
    if !fr.ran { t.Fatalf("expected runner.Run to be called") }
    if calledFatal { t.Fatalf("did not expect fatalf to be called") }
}

func TestRun_FatalOnCtorError(t *testing.T) {
    oldCtor := appCtor
    oldFatalf := fatalf
    t.Cleanup(func(){ appCtor = oldCtor; fatalf = oldFatalf })

    appCtor = func() (runner, error) { return nil, errors.New("boom") }

    calledFatal := false
    fatalf = func(format string, v ...any) { calledFatal = true }

    run(context.Background())

    if !calledFatal { t.Fatalf("expected fatalf to be called on ctor error") }
}

func TestRun_FatalOnRunError(t *testing.T) {
    fr := &fakeRunner{err: errors.New("oops")}
    oldCtor := appCtor
    oldFatalf := fatalf
    t.Cleanup(func(){ appCtor = oldCtor; fatalf = oldFatalf })

    appCtor = func() (runner, error) { return fr, nil }

    calledFatal := false
    fatalf = func(format string, v ...any) { calledFatal = true }

    run(context.Background())

    if !calledFatal { t.Fatalf("expected fatalf to be called on run error") }
}
