import logic
section
  variables (A : Type) (p q : A → Prop)

  example : (∀x : A, p x ∧ q x) → ∀y : A, p y  :=
  assume H : ∀x : A, p x ∧ q x,
  take y : A,
  show p y, from and.elim_left (H y)
end
