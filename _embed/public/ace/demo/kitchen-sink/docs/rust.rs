use core::rand::RngUtil;

fn main() {
    for ["Alice", "Bob", "Carol"].each |&name| {
        do spawn {
            let v = rand::Rng().shuffle([1, 2, 3]);
            for v.each |&num| {
                print(fmt!("%s says: '%d'\n", name, num + 1))
            }
        }
    }
}

fn map<T, U>(vector: &[T], function: &fn(v: &T) -> U) -> ~[U] {
    let mut accumulator = ~[];
    for vec::each(vector) |element| {
        accumulator.push(function(element));
    }
    return accumulator;
}
