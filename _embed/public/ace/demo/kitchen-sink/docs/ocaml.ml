(*
 * Example of early return implementation taken from
 * http://ocaml.janestreet.com/?q=node/91
 *)

let with_return (type t) (f : _ -> t) =
  let module M =
     struct exception Return of t end
  in
  let return = { return = (fun x -> raise (M.Return x)); } in
  try f return with M.Return x -> x


(* Function that uses the 'early return' functionality provided by `with_return` *)
let sum_until_first_negative list =
  with_return (fun r ->
    List.fold list ~init:0 ~f:(fun acc x ->
      if x >= 0 then acc + x else r.return acc))