package main

type gameObject interface {
    init();
    update(int64, *Game);
    destroy();
}
