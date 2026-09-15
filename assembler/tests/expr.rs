use corewar_asm::expr::parse_expression;
use std::collections::HashMap;

#[test]
fn expression_precedence_and_labels() {
    let mut labels = HashMap::new();
    labels.insert("x".to_string(), 20usize);
    assert_eq!(
        parse_expression("1 + 2 * 3")
            .unwrap()
            .eval(&labels, 0)
            .unwrap(),
        7
    );
    assert_eq!(
        parse_expression("(:x - 2) / 3")
            .unwrap()
            .eval(&labels, 5)
            .unwrap(),
        13 / 3
    );
    assert_eq!(
        parse_expression("-(4 + 2)")
            .unwrap()
            .eval(&labels, 0)
            .unwrap(),
        -6
    );
}
