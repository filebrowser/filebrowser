// ancestors
parentof("douglas", "john").
parentof("john", "bob").
parentof("bob", "ebbon").

parentof("douglas", "jane").
parentof("jane", "jan").

ancestorof(A, B) <- parentof(A, B).
ancestorof(A, C) <- ancestorof(A, B), parentof(B,C).

grandparentof(A, B) <- parentof(A, C), parentof(C, B).

cousins(A,B) <- grandparentof(C,A), grandparentof(C,B).

parentof[`arg](A, B) -> int[32](A), !string(B).