import os


def extract_parsed_literal_expression_nodes(node):
    literal_expression_nodes = []

    if "type" not in node:
        return []

    if node["type"] == "parsed_literal_expression":
        literal_expression_nodes.append(node)
    elif node["type"] in ["and", "or"]:
        literal_expression_nodes += extract_parsed_literal_expression_nodes(
            node["right"]
        )
        literal_expression_nodes += extract_parsed_literal_expression_nodes(
            node["left"]
        )

    return literal_expression_nodes


def check_vulnerable(request, results_path):
    param_values = [param["value"] for param in request.params.values()]

    for query in request.queries:
        if query.ast is None:
            continue
        literal_expression_nodes = extract_parsed_literal_expression_nodes(
            query.ast["where_clause"]
        )
        for node in literal_expression_nodes:
            # if any(str(value).upper() in node["literal"].upper() for value in param_values):
            for value in param_values:
                if str(value).upper() in node["literal"].upper():
                    with open(os.path.join(results_path, "requests.log"), "a") as f:
                        f.write("\nsql injecton:\n")
                        f.write("param: %s\n" % value)
                        f.write("literal expression node: {}\n".format(node["literal"]))
                        f.write("query: {}\n\n".format(query.ast["where_clause"]))
